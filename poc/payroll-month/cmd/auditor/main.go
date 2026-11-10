// Command auditor is the independent verifier (prompt8 §5.2/§5.3). It
// holds its OWN trusted state (separate dir/volume from the writer),
// connects to immudb as a READ-ONLY user and to PostgreSQL through
// a SELECT-only role, and verifies ledger evidence against
// independently reconstructed payroll content.
//
// Modes:
//   - Content (default): --employer CODE --run UUID. Rebuilds the
//     expected evidence from authoritative PG sources, re-executes
//     the payroll from persisted inputs, and compares everything
//     against cryptographically verified ledger bytes. Exit 0 only
//     when every event verifies.
//   - Liveness: --txs 1,2,3. Verifies tx inclusion only. Labeled as
//     such in the report: liveness is NOT content verification.
//
// Genesis (§5.3): the first run against an empty state dir requires
// --accept-genesis and prints the committable anchor (server UUID +
// content hash of tx 1, the immutable history root) to stderr.
// Ordinary runs pin the stack with
// --expect-uuid/--expect-tx/--expect-tx-hash (all three or none): the
// pin is checked in a fresh temp dir BEFORE any trust decision, and
// any mismatch exits 2. The anchor for this stack is committed at
// gate/immudb-genesis.json (genesis ceremony, runbook step).
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PithomLabs/a2/poc/payroll-month/ledger"
)

type txResult struct {
	Tx       uint64 `json:"tx"`
	Verified bool   `json:"verified"`
	Error    string `json:"error,omitempty"`
}

type report struct {
	Mode      string        `json:"mode"`
	Genesis   bool          `json:"genesis_accepted"`
	Pinned    bool          `json:"pin_checked"`
	ServerID  string        `json:"server_uuid,omitempty"`
	StateHash string        `json:"state_hash"`
	Employer  string        `json:"employer,omitempty"`
	Run       string        `json:"run,omitempty"`
	Results   []txResult    `json:"tx_results,omitempty"`
	Events    []eventResult `json:"events,omitempty"`
	Global    []checkResult `json:"global_checks,omitempty"`
	Overall   string        `json:"overall"`
}

type anchor struct {
	ServerUUID string `json:"server_uuid"`
	AnchorTx   uint64 `json:"anchor_tx"`
	TxHash     string `json:"tx_hash"`
}

func main() { os.Exit(run()) }

func run() int {
	addr := flag.String("addr", getenv("POC_IMMUDB_ADDR", "localhost:3322"), "immudb host:port")
	user := flag.String("user", getenv("POC_AUDIT_IMMUDB_USER", "poc_auditor"), "immudb user (read-only)")
	pass := flag.String("pass", os.Getenv("POC_AUDIT_IMMUDB_PASSWORD"), "immudb password")
	db := flag.String("db", getenv("POC_IMMUDB_DB", "defaultdb"), "immudb database")
	pgURL := flag.String("pg", os.Getenv("POC_AUDIT_DATABASE_URL"), "PostgreSQL URL (SELECT-only role)")
	stateDir := flag.String("state-dir", os.Getenv("POC_AUDIT_STATE_DIR"), "OWN state dir (never the writer's)")
	employer := flag.String("employer", "", "employer code (content mode)")
	runID := flag.String("run", "", "payroll_run id (content mode)")
	txs := flag.String("txs", "", "comma-separated tx ids (liveness mode only)")
	acceptGenesis := flag.Bool("accept-genesis", false, "required on first run with empty state dir")
	expectUUID := flag.String("expect-uuid", "", "pinned server UUID (with --expect-tx/--expect-tx-hash)")
	expectTx := flag.Uint64("expect-tx", 0, "pinned anchor tx (with --expect-uuid/--expect-tx-hash)")
	expectHash := flag.String("expect-tx-hash", "", "pinned anchor-tx content hash (with --expect-uuid/--expect-tx)")
	webhook := flag.String("webhook", os.Getenv("POC_AUDIT_WEBHOOK"), "optional failure webhook URL")
	flag.Parse()

	if *stateDir == "" {
		fmt.Fprintln(os.Stderr, "auditor: --state-dir is required (own volume, never shared)")
		return 2
	}
	// The auditor must never run on the writer's trusted state: same
	// dir means shared fate with the party under audit.
	if wdir := os.Getenv("POC_WRITER_STATE_DIR"); wdir != "" {
		a, aerr := filepath.Abs(*stateDir)
		w, werr := filepath.Abs(wdir)
		if aerr == nil && werr == nil && a == w {
			fmt.Fprintln(os.Stderr, "auditor: --state-dir equals POC_WRITER_STATE_DIR: refusing to share trusted state with the writer")
			return 2
		}
	}
	content := *runID != ""
	if content == (*txs != "") {
		fmt.Fprintln(os.Stderr, "auditor: pass exactly one of --run (content) or --txs (liveness)")
		return 2
	}
	if content && *employer == "" {
		fmt.Fprintln(os.Stderr, "auditor: content mode requires --employer")
		return 2
	}
	pinCount := 0
	for _, s := range []string{*expectUUID, *expectHash} {
		if s != "" {
			pinCount++
		}
	}
	if *expectTx != 0 {
		pinCount++
	}
	if pinCount != 0 && pinCount != 3 {
		fmt.Fprintln(os.Stderr, "auditor: --expect-uuid/--expect-tx/--expect-tx-hash must be given together")
		return 2
	}
	if content && *pgURL == "" {
		fmt.Fprintln(os.Stderr, "auditor: content mode requires --pg (SELECT-only role URL)")
		return 2
	}
	empty, err := dirIsEmpty(*stateDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "auditor: state dir:", err)
		return 2
	}
	if empty && !*acceptGenesis {
		fmt.Fprintln(os.Stderr, "auditor: empty state dir without --accept-genesis: refusing to bootstrap trust implicitly")
		return 2
	}
	if !empty && *acceptGenesis {
		fmt.Fprintln(os.Stderr, "auditor: --accept-genesis on a non-empty state dir: refusing to re-bootstrap trust (wipe explicitly for a new ceremony)")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	host, port := splitAddr(*addr)
	rep := report{Genesis: empty && *acceptGenesis}

	// Pin check first (fresh temp dir, before any trust decision).
	if pinCount == 3 {
		if err := checkPin(ctx, host, port, *user, *pass, *db, *expectUUID, *expectTx, *expectHash); err != nil {
			fmt.Fprintln(os.Stderr, "auditor: genesis pin:", err)
			return 2
		}
		rep.Pinned = true
		rep.ServerID = *expectUUID
	}

	app, err := ledger.DialDir(ctx, host, port, *user, *pass, *db, *stateDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "auditor: dial:", err)
		return 1
	}
	defer app.Close(ctx)

	ok := true
	if !content {
		rep.Mode = "liveness (tx-existence only, NOT content verification)"
		ids, err := resolveTxIDs(*txs)
		if err != nil {
			fmt.Fprintln(os.Stderr, "auditor:", err)
			return 2
		}
		for _, id := range ids {
			tr := txResult{Tx: id, Verified: true}
			if err := app.VerifiedTx(ctx, id); err != nil {
				tr.Verified = false
				tr.Error = err.Error()
				ok = false
			}
			rep.Results = append(rep.Results, tr)
		}
	} else {
		rep.Mode = "content"
		rep.Employer, rep.Run = *employer, *runID
		pool, err := pgxpool.New(ctx, *pgURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, "auditor: pg:", err)
			return 1
		}
		defer pool.Close()
		events, global, err := verifyContent(ctx, contentInput{
			pool: pool, app: app, employer: *employer, runID: *runID,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "auditor: content:", err)
			return 1
		}
		rep.Events, rep.Global = events, global
		for _, g := range global {
			if !g.Pass {
				ok = false
			}
		}
		for _, e := range events {
			if !e.Verified {
				ok = false
			}
		}
	}
	uuid, err := stateServerUUID(*stateDir)
	if err == nil {
		rep.ServerID = uuid
	}
	rep.StateHash, err = hashStateDir(*stateDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "auditor: state hash:", err)
		return 1
	}
	rep.Overall = map[bool]string{true: "pass", false: "fail"}[ok]
	raw, _ := json.MarshalIndent(rep, "", "  ")
	fmt.Println(string(raw))
	if rep.Genesis {
		anch, err := ceremonyAnchor(ctx, host, port, *user, *pass, *db)
		if err != nil {
			fmt.Fprintln(os.Stderr, "auditor: GENESIS accepted but anchor capture failed:", err)
			return 1
		}
		araw, _ := json.Marshal(anch)
		fmt.Fprintf(os.Stderr, "auditor: GENESIS accepted — commit this anchor out-of-band (git, signed): %s\n", araw)
	}
	if !ok && *webhook != "" {
		postWebhook(*webhook, raw)
	}
	if !ok {
		return 1
	}
	return 0
}

// checkPin verifies the stack identity in a fresh temp dir: verify the
// anchor tx from scratch, then compare the observed server UUID and
// anchor-tx content hash against the pinned values. Any mismatch
// exits loudly before the main state dir is touched.
func checkPin(ctx context.Context, host string, port int, user, pass, db, uuid string, tx uint64, hash string) error {
	dir, err := os.MkdirTemp("", "audit-pin-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	app, err := ledger.DialDir(ctx, host, port, user, pass, db, dir)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer app.Close(ctx)
	gotHash, err := app.VerifiedTxHash(ctx, tx)
	if err != nil {
		return fmt.Errorf("anchor tx %d: %w", tx, err)
	}
	gotUUID, err := stateServerUUID(dir)
	if err != nil {
		return fmt.Errorf("uuid: %w", err)
	}
	if gotUUID != uuid {
		return fmt.Errorf("server UUID %q != pinned %q", gotUUID, uuid)
	}
	if gotHash != hash {
		return fmt.Errorf("anchor tx hash %q != pinned %q", gotHash, hash)
	}
	return nil
}

// ceremonyAnchor captures the committable anchor: the server UUID
// plus the content hash of tx 1 (the immutable history root),
// verified from a fresh dir. Stable for the life of the volume.
func ceremonyAnchor(ctx context.Context, host string, port int, user, pass, db string) (anchor, error) {
	var a anchor
	a.AnchorTx = 1
	dir, err := os.MkdirTemp("", "audit-genesis-")
	if err != nil {
		return a, err
	}
	defer os.RemoveAll(dir)
	app, err := ledger.DialDir(ctx, host, port, user, pass, db, dir)
	if err != nil {
		return a, fmt.Errorf("dial: %w", err)
	}
	defer app.Close(ctx)
	a.TxHash, err = app.VerifiedTxHash(ctx, 1)
	if err != nil {
		return a, fmt.Errorf("anchor tx 1: %w", err)
	}
	a.ServerUUID, err = stateServerUUID(dir)
	if err != nil {
		return a, err
	}
	return a, nil
}

// stateServerUUID reads the server UUID the client pinned in its
// identity file (format: {"serverUUID": ..., "serverIdentity": ...}).
func stateServerUUID(dir string) (string, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range ents {
		if e.IsDir() || !strings.HasPrefix(e.Name(), ".identity-") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return "", err
		}
		var v struct {
			ServerUUID string `json:"serverUUID"`
		}
		if err := json.Unmarshal(raw, &v); err != nil {
			return "", fmt.Errorf("identity file: %w", err)
		}
		if v.ServerUUID == "" {
			return "", fmt.Errorf("identity file has no serverUUID")
		}
		return v.ServerUUID, nil
	}
	return "", fmt.Errorf("no identity file in %s", dir)
}

func resolveTxIDs(list string) ([]uint64, error) {
	var out []uint64
	for _, s := range strings.Split(list, ",") {
		n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad tx id %q", s)
		}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no tx ids: pass --txs")
	}
	return out, nil
}

func dirIsEmpty(dir string) (bool, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return false, err
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(ents) == 0, nil
}

// hashStateDir hashes (relative path, bytes) over the state dir. Paths
// must be relative: absolute paths would make the hash dir-dependent
// and no committed anchor could ever match (found by probe).
func hashStateDir(dir string) (string, error) {
	var files []string
	if err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return "", err
		}
		h.Write([]byte(f))
		h.Write(raw)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func postWebhook(url string, body []byte) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "auditor: webhook build:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "auditor: webhook post:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Fprintf(os.Stderr, "auditor: webhook status %s\n", resp.Status)
}

func splitAddr(addr string) (string, int) {
	h, p, _ := strings.Cut(addr, ":")
	port, err := strconv.Atoi(p)
	if err != nil {
		return h, 3322
	}
	return h, port
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
