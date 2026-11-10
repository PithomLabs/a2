// Package ledger appends tamper-evident provenance records to a standalone
// immudb server over gRPC. Records carry hashes and metadata ONLY — never
// PII, secrets, or erasure-subject data (immudb deletes are logical-only).
package ledger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

// Record is the canonical provenance value stored under prov/... keys
// (prompt8 §5.1). No post-commit field (ledger tx id, ALH) lives in the
// payload — those are PG back-links (see outbox.sql) — so the bytes
// hashed are exactly the bytes stored. Corrections are new events
// chained by Supersedes, never overwrites.
//
// Field provenance (each independently re-derivable by the auditor
// from PostgreSQL — see cmd/auditor):
//   - InputDigest: tip payroll_snapshot.content_hash at enqueue.
//   - OutputDigest: poc_run_output.output_digest (sha256 over the
//     canonical line segments, see poc_result_persist).
//   - SHA256: hex sha256 over the canonical outbox payload bytes
//     (payload_hash); binds the record to its exact enqueue bytes.
//   - RuleVersions/PackDigest: snapshot pins (pack->version;
//     PackDigest is the FIRST pin's manifest digest in
//     (effective_from, pack) order).
//   - CalcDigest: sha256 over the pinned calc_version LABEL (binds
//     the version, not implementation bytes — implementation
//     provenance is future work, honestly).
//   - DecidedBy/DecidedAt: the approve decision (POC operator label).
//
// No PII, amounts, names, NRICs, bank details, or secrets: digests,
// UUIDs, codes, and the operator label only.
type Record struct {
	EventUUID    string            `json:"event_uuid"`
	Artifact     string            `json:"artifact"`
	ArtifactID   string            `json:"artifact_uuid"`
	PGTable      string            `json:"pg_table"`
	PGID         string            `json:"pg_id"`
	EmployerID   string            `json:"employer_id"`
	SHA256       string            `json:"sha256"`
	InputDigest  string            `json:"input_digest"`
	OutputDigest string            `json:"output_digest"`
	RuleVersions map[string]string `json:"rule_versions"`
	PackDigest   string            `json:"rule_pack_digest"`
	CalcDigest   string            `json:"calc_digest"`
	DecidedBy    string            `json:"decided_by"`
	DecidedAt    string            `json:"decided_at"`
	Supersedes   string            `json:"supersedes"` // prev event_uuid, "" = none
}

// CanonicalBytes is the digest input. encoding/json sorts map keys, keeping
// this deterministic.
func (r Record) CanonicalBytes() []byte {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil
	}
	return raw
}

// Digest returns the hex sha256 of the canonical bytes.
func (r Record) Digest() string {
	sum := sha256.Sum256(r.CanonicalBytes())
	return hex.EncodeToString(sum[:])
}

// ProvKey builds prov/{employer}/{artifact}/{uuid}/{event} ledger keys.
// Every evidence event is a unique key: retries reuse the SAME event uuid,
// corrections mint a new one.
func ProvKey(employer, artifact, id, event string) []byte {
	return []byte("prov/" + employer + "/" + artifact + "/" + id + "/" + event)
}

// IdemKey builds idem/{employer}/{action}/{uuid} idempotency keys. The same
// business op_uuid is shared with the poc_outbox row (see outbox.sql).
func IdemKey(employer, action, uuid string) []byte {
	return []byte("idem/" + employer + "/" + action + "/" + uuid)
}

// Appender is the single-gated ledger writer: all POC appends flow through
// one Appender bound to one immudb user/database.
type Appender struct {
	cli client.ImmuClient
}

// Dial connects and opens a session with the default (process-local) state
// dir. Prefer DialDir with an explicit volume-backed dir.
func Dial(ctx context.Context, address string, port int, user, pass, database string) (*Appender, error) {
	return DialDir(ctx, address, port, user, pass, database, ".")
}

// DialDir connects with an explicit file-cache dir for trusted state. The
// writer's dir and the auditor's dir must be DIFFERENT volumes (R2 model);
// insecure gRPC is POC-local only.
func DialDir(ctx context.Context, address string, port int, user, pass, database, dir string) (*Appender, error) {
	// Session flow (NOT NewImmuClient, which connects immediately and
	// makes the later OpenSession fail with "session already opened" —
	// proven at Gate B). Defaults first: bare Options zeroes
	// StreamChunkSize, which OpenSession rejects ("minimum chunk size
	// is 4096").
	opts := client.DefaultOptions()
	opts.Address, opts.Port, opts.Dir = address, port, dir
	opts.DialOptions = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	// The session flow (unlike NewImmuClient) never creates the state
	// dir; an absent dir fails verified ops with "no such file".
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("ledger: state dir: %w", err)
	}
	cli := client.NewClient().WithOptions(opts)
	if err := cli.OpenSession(ctx, []byte(user), []byte(pass), database); err != nil {
		return nil, fmt.Errorf("ledger: open session: %w", err)
	}
	return &Appender{cli: cli}, nil
}

// Close ends the session.
func (a *Appender) Close(ctx context.Context) error {
	return a.cli.CloseSession(ctx)
}

// Append writes the provenance record plus its idempotency marker in one
// atomic ExecAll guarded by must-not-exist preconditions on BOTH keys, then
// verifies the commit with dual-proof verification and returns the immudb tx
// id. The marker value binds event_uuid + content digest: a precondition
// failure means this op_uuid was already appended, and the caller must
// re-read the marker through a VERIFIED read — equal digest recovers the
// origin tx id and continues to the back-link INSERT; different digest
// fails loudly (payload aliasing). Neither namespace is ever deleted or
// expired (see ADR-0204).
func (a *Appender) Append(ctx context.Context, rec Record, idemKey []byte) (uint64, error) {
	val, err := json.Marshal(rec)
	if err != nil {
		return 0, fmt.Errorf("ledger: marshal record: %w", err)
	}
	if rec.EventUUID == "" {
		return 0, fmt.Errorf("ledger: empty event_uuid")
	}
	provKey := ProvKey(rec.EmployerID, rec.Artifact, rec.ArtifactID, rec.EventUUID)
	marker := rec.EventUUID + ":" + rec.Digest()
	req := &schema.ExecAllRequest{
		Operations: []*schema.Op{
			{Operation: &schema.Op_Kv{Kv: &schema.KeyValue{
				Key:   provKey,
				Value: val,
			}}},
			{Operation: &schema.Op_Kv{Kv: &schema.KeyValue{
				Key:   idemKey,
				Value: []byte(marker),
			}}},
		},
		Preconditions: []*schema.Precondition{
			schema.PreconditionKeyMustNotExist(provKey),
			schema.PreconditionKeyMustNotExist(idemKey),
		},
		NoWait: false, // never trade freshness on the append path
	}
	hdr, err := a.cli.ExecAll(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("ledger: execall: %w", err)
	}
	if hdr == nil || hdr.Id == 0 {
		return 0, fmt.Errorf("ledger: empty tx header")
	}
	if _, err := a.cli.VerifiedTxByID(ctx, hdr.Id); err != nil {
		return 0, fmt.Errorf("ledger: verify tx %d: %w", hdr.Id, err)
	}
	return hdr.Id, nil
}

// ReadMarker returns the idempotency marker (value, origin tx id). It
// uses a plain Get: dial-tone reads and test probes only. Every
// recovery/back-link decision MUST use VerifiedReadMarker instead
// (prompt8 §5.4: never link from an unverified server response).
func (a *Appender) ReadMarker(ctx context.Context, key []byte) (string, uint64, error) {
	e, err := a.cli.Get(ctx, key)
	if err != nil {
		return "", 0, fmt.Errorf("ledger: read marker: %w", err)
	}
	return string(e.Value), e.Tx, nil
}

// VerifiedReadMarker returns the marker through dual-proof verification
// (inclusion proof + state advancement against this client's trusted
// state). The origin tx id comes from the verified entry itself, so
// recovery needs no scan. Idem keys are never rewritten, so the
// current version is the only version.
func (a *Appender) VerifiedReadMarker(ctx context.Context, key []byte) (string, uint64, error) {
	e, err := a.cli.VerifiedGet(ctx, key)
	if err != nil {
		return "", 0, fmt.Errorf("ledger: verified marker read: %w", err)
	}
	return string(e.Value), e.Tx, nil
}

// VerifiedTx re-verifies a historical tx against trusted state (dual
// proof). Recovery and the auditor use it to confirm an anchored tx
// is still part of the server's attested history.
func (a *Appender) VerifiedTx(ctx context.Context, tx uint64) error {
	if _, err := a.cli.VerifiedTxByID(ctx, tx); err != nil {
		return fmt.Errorf("ledger: verify tx %d: %w", tx, err)
	}
	return nil
}

// VerifiedTxHash returns the content hash (sha256 over deterministic
// protobuf bytes) of a verified historical tx. The genesis pin
// anchors tx 1's hash: the server UUID pins the instance, the tx
// hash pins the history root. (A state-dir hash cannot serve this
// purpose: every verified read advances trusted state to the tip,
// so it stales on each append — proven live.)
func (a *Appender) VerifiedTxHash(ctx context.Context, tx uint64) (string, error) {
	verified, err := a.cli.VerifiedTxByID(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("ledger: verify tx %d: %w", tx, err)
	}
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(verified)
	if err != nil {
		return "", fmt.Errorf("ledger: marshal tx %d: %w", tx, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// VerifiedReadRecord returns the provenance record stored under key
// through dual-proof verification, plus its origin tx. Prov keys are
// never rewritten (MustNotExist precondition), so the current
// version is the appended version.
func (a *Appender) VerifiedReadRecord(ctx context.Context, key []byte) (Record, uint64, error) {
	var rec Record
	e, err := a.cli.VerifiedGet(ctx, key)
	if err != nil {
		return rec, 0, fmt.Errorf("ledger: verified record read: %w", err)
	}
	if err := json.Unmarshal(e.Value, &rec); err != nil {
		return rec, 0, fmt.Errorf("ledger: record decode: %w", err)
	}
	return rec, e.Tx, nil
}

// VerifiedReadRecordAt returns the record value as of tx (dual proof).
// Recovery uses it to bind the marker's origin tx to the exact record
// bytes that tx committed (marker and record append atomically).
func (a *Appender) VerifiedReadRecordAt(ctx context.Context, key []byte, tx uint64) (Record, error) {
	var rec Record
	e, err := a.cli.VerifiedGetAt(ctx, key, tx)
	if err != nil {
		return rec, fmt.Errorf("ledger: verified record read at %d: %w", tx, err)
	}
	if err := json.Unmarshal(e.Value, &rec); err != nil {
		return rec, fmt.Errorf("ledger: record decode: %w", err)
	}
	return rec, nil
}
