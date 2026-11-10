// Content verification (prompt8 §5.2): the auditor reads PostgreSQL
// through a read-only role, rebuilds the expected evidence from
// authoritative sources, re-executes the payroll from persisted
// inputs, and compares everything against cryptographically verified
// ledger bytes. It never trusts a supplied backlinks file (there is
// no such input): every binding is re-derived.
//
// Verification equation per evidence event:
//
//	A. stored output_digest == ledger OutputDigest (output binding)
//	B. stored tip content_hash == ledger InputDigest (input binding)
//	C. re-executed lines == persisted payroll_line rows, field by
//	   field (inputs genuinely produce the persisted outputs; detects
//	   input AND output tamper)
//	D. canon digest over PERSISTED rows == ledger OutputDigest
//	   (pure PG-side recompute; detects output tamper alone)
//	E. snapshot pins == ledger RuleVersions/PackDigest
//	F. sha256(stored calc_version) == ledger CalcDigest
//	G. stored approval == ledger DecidedBy/DecidedAt
//	H. recomputed supersedes predecessor == ledger Supersedes
//	I. sha256(stored payload bytes) == ledger SHA256
//	J. identity: ledger event/artifact/table/id/employer == the
//	   PG-claimed binding; marker verified with origin == backlink
//	   tx; record origin == backlink tx; received bytes ==
//	   reconstructed bytes
//	K. run status is validated (drafts are refused, never attested)
//	L. every backlink for the run verifies; every terminal outbox
//	   row has its backlink (no missing links)
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PithomLabs/a2/poc/payroll-month/dbosflow"
	"github.com/PithomLabs/a2/poc/payroll-month/ledger"
)

type checkResult struct {
	Name   string `json:"name"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail,omitempty"`
}

type eventResult struct {
	EventUUID string        `json:"event_uuid"`
	Tx        uint64        `json:"tx"`
	Verified  bool          `json:"verified"`
	Checks    []checkResult `json:"checks"`
}

// contentInput carries the verified endpoints for one audit run.
type contentInput struct {
	pool     *pgxpool.Pool
	app      *ledger.Appender
	employer string // employer code
	runID    string // payroll_run id (any case; normalized)
}

// verifyContent runs checks A–L for every evidence event of the run.
func verifyContent(ctx context.Context, in contentInput) ([]eventResult, []checkResult, error) {
	run := strings.ToLower(in.runID)
	var global []checkResult
	fail := func(name, detail string) {
		global = append(global, checkResult{Name: name, Detail: detail})
	}

	// K. The run exists, belongs to the employer, and is validated.
	var employerID, status, calcVersion string
	err := in.pool.QueryRow(ctx,
		`SELECT r.employer_id::text, r.status, r.calc_version
		   FROM payroll_run r JOIN employer e ON e.id = r.employer_id
		  WHERE r.id = $1 AND e.code = $2`, run, in.employer).Scan(&employerID, &status, &calcVersion)
	if err != nil {
		fail("K.run-finalized", fmt.Sprintf("run %s for %s: %v", run, in.employer, err))
		return nil, global, nil
	}
	if status != "validated" {
		fail("K.run-finalized", fmt.Sprintf("run %s status %q, want validated", run, status))
		return nil, global, nil
	}
	global = append(global, checkResult{Name: "K.run-finalized", Pass: true})
	var payDate string
	if err := in.pool.QueryRow(ctx,
		`SELECT p.pay_date::text FROM pay_period p
		  JOIN payroll_run r ON r.pay_period_id = p.id WHERE r.id = $1`, run).Scan(&payDate); err != nil {
		fail("K.run-period", err.Error())
		return nil, global, nil
	}

	// Sources shared by every event.
	var tipID, tipHash string
	if err := in.pool.QueryRow(ctx,
		`SELECT id, content_hash FROM payroll_snapshot
		  WHERE payroll_run_id = $1 AND status = 'valid'`, run).Scan(&tipID, &tipHash); err != nil {
		fail("B.tip", err.Error())
		return nil, global, nil
	}
	var storedOut, storedIn string
	if err := in.pool.QueryRow(ctx,
		`SELECT output_digest, input_digest FROM poc_run_output WHERE run_id = $1`,
		run).Scan(&storedOut, &storedIn); err != nil {
		fail("A.output-row", err.Error())
		return nil, global, nil
	}
	// The persist bound its input at persist time; a drift between the
	// bound input and the current tip is itself evidence of tamper.
	if storedIn != tipHash {
		fail("B.persist-bind", fmt.Sprintf("persisted input %q != tip %q", trunc(storedIn), trunc(tipHash)))
		return nil, global, nil
	}
	global = append(global, checkResult{Name: "B.persist-bind", Pass: true})
	type pin struct{ pack, version, digest string }
	var pins []pin
	prows, err := in.pool.Query(ctx,
		`SELECT pack, version, manifest_digest FROM payroll_snapshot_pack
		  WHERE snapshot_id = $1 ORDER BY effective_from, pack`, tipID)
	if err != nil {
		fail("E.pins", err.Error())
		return nil, global, nil
	}
	for prows.Next() {
		var p pin
		if err := prows.Scan(&p.pack, &p.version, &p.digest); err != nil {
			prows.Close()
			fail("E.pins", err.Error())
			return nil, global, nil
		}
		pins = append(pins, p)
	}
	prows.Close()
	if err := prows.Err(); err != nil {
		fail("E.pins", err.Error())
		return nil, global, nil
	}
	var decidedBy, decidedAt string
	if err := in.pool.QueryRow(ctx,
		`SELECT approver, decided_at::text FROM poc_approval
		  WHERE run_id = $1 AND decision = 'approve'`, run).Scan(&decidedBy, &decidedAt); err != nil {
		fail("G.approval", err.Error())
		return nil, global, nil
	}

	// C. Re-execute from persisted inputs (shared pure mapping,
	// independent data): recomputed rows must field-match persisted.
	members, snapPins, err := dbosflow.LoadSnapshot(ctx, in.pool, run)
	if err != nil {
		fail("C.reload", err.Error())
		return nil, global, nil
	}
	packs, err := dbosflow.PacksToRulePacks(snapPins)
	if err != nil {
		fail("C.packs", err.Error())
		return nil, global, nil
	}
	m, _, err := dbosflow.CalcMonth(members, packs, "2026-10", in.employer)
	if err != nil {
		fail("C.recompute", err.Error())
		return nil, global, nil
	}
	wantLines, wantSlips := dbosflow.BuildPersistLines(members, snapPins, m, payDate)

	type pgLine struct {
		no, comp, code, cents, hours, cpf, wp, pd, cp, refs string
	}
	var pgLines []pgLine
	lrows, err := in.pool.Query(ctx,
		`SELECT m.employee_no, l.component, l.pay_code, l.amount::text,
		        l.hours::text, l.cpf_treatment, l.work_period,
		        l.payable_date::text, l.contribution_period, l.source_refs::text
		   FROM payroll_line l JOIN employee m ON m.id = l.employee_id
		  WHERE l.run_id = $1
		  ORDER BY m.employee_no,
		    CASE l.component WHEN 'basic' THEN 1 WHEN 'overtime' THEN 2
		         WHEN 'leave' THEN 3 WHEN 'allowance' THEN 4
		         WHEN 'deduction' THEN 5 ELSE 99 END,
		    l.pay_code`, run)
	if err != nil {
		fail("C.lines", err.Error())
		return nil, global, nil
	}
	for lrows.Next() {
		var l pgLine
		var hours *string
		if err := lrows.Scan(&l.no, &l.comp, &l.code, &l.cents, &hours, &l.cpf, &l.wp, &l.pd, &l.cp, &l.refs); err != nil {
			lrows.Close()
			fail("C.lines", err.Error())
			return nil, global, nil
		}
		if hours != nil {
			l.hours = *hours
		}
		pgLines = append(pgLines, l)
	}
	lrows.Close()
	if err := lrows.Err(); err != nil {
		fail("C.lines", err.Error())
		return nil, global, nil
	}
	if len(pgLines) != len(wantLines) {
		fail("C.lines", fmt.Sprintf("persisted %d lines, recomputed %d", len(pgLines), len(wantLines)))
		return nil, global, nil
	}
	for i := range wantLines {
		w, g := wantLines[i], pgLines[i]
		gotCents, err := parseCents(g.cents)
		if err != nil {
			fail("C.lines", fmt.Sprintf("row %d unparsable amount %q: %v", i, g.cents, err))
			return nil, global, nil
		}
		if w.EmployeeNo != g.no || w.Component != g.comp || w.PayCode != g.code || w.AmountCents != gotCents {
			fail("C.lines", fmt.Sprintf("row %d: persisted %s/%s/%s/%d != recomputed %s/%s/%s/%d",
				i, g.no, g.comp, g.code, gotCents, w.EmployeeNo, w.Component, w.PayCode, w.AmountCents))
			return nil, global, nil
		}
		if (w.Hours == nil) != (g.hours == "") {
			fail("C.lines", fmt.Sprintf("row %d hours diverged", i))
			return nil, global, nil
		}
		if w.CPFTreatment != g.cpf || w.WorkPeriod != g.wp || w.PayableDate != g.pd || w.ContributionPeriod != g.cp {
			fail("C.lines", fmt.Sprintf("row %d treatment/period diverged", i))
			return nil, global, nil
		}
		if !jsonEqual(w.SourceRefs, g.refs) {
			fail("C.lines", fmt.Sprintf("row %d source_refs diverged", i))
			return nil, global, nil
		}
	}
	global = append(global, checkResult{Name: "C.lines", Pass: true,
		Detail: fmt.Sprintf("%d rows field-match recomputation", len(wantLines))})

	// Payslips: recomputed payloads must match persisted payloads.
	gotSlips := map[string]string{}
	srows, err := in.pool.Query(ctx,
		`SELECT m.employee_no, p.payload::text FROM payslip p
		  JOIN employee m ON m.id = p.employee_id WHERE p.run_id = $1`, run)
	if err != nil {
		fail("C.slips", err.Error())
		return nil, global, nil
	}
	for srows.Next() {
		var no, payload string
		if err := srows.Scan(&no, &payload); err != nil {
			srows.Close()
			fail("C.slips", err.Error())
			return nil, global, nil
		}
		gotSlips[no] = payload
	}
	srows.Close()
	if err := srows.Err(); err != nil {
		fail("C.slips", err.Error())
		return nil, global, nil
	}
	if len(gotSlips) != len(wantSlips) {
		fail("C.slips", fmt.Sprintf("persisted %d slips, recomputed %d", len(gotSlips), len(wantSlips)))
		return nil, global, nil
	}
	for _, w := range wantSlips {
		g, ok := gotSlips[w.EmployeeNo]
		if !ok || !jsonEqual(w.Payload, g) {
			fail("C.slips", fmt.Sprintf("slip %s diverged", w.EmployeeNo))
			return nil, global, nil
		}
	}
	global = append(global, checkResult{Name: "C.slips", Pass: true,
		Detail: fmt.Sprintf("%d payloads match", len(wantSlips))})

	// D. Canon digest over PERSISTED rows (pure PG-side recompute).
	var canon strings.Builder
	for _, l := range pgLines {
		cents, err := parseCents(l.cents)
		if err != nil {
			fail("D.canon", err.Error())
			return nil, global, nil
		}
		fmt.Fprintf(&canon, "%s|%s|%s|%d;", l.no, l.comp, l.code, cents)
	}
	canonSum := sha256.Sum256([]byte("output|" + run + "|" + canon.String()))
	canonDigest := hex.EncodeToString(canonSum[:])

	// L (coverage): terminal outbox rows for this artifact.
	type opRow struct {
		action, state, payloadText string
		hash                       []byte
	}
	ops := map[string]opRow{}
	orows, err := in.pool.Query(ctx,
		`SELECT op_uuid::text, action, state, payload::text, payload_hash
		   FROM poc_outbox
		  WHERE payload->>'pg_id' = $1 AND action = 'payroll.finalized'`, run)
	if err != nil {
		fail("L.outbox", err.Error())
		return nil, global, nil
	}
	for orows.Next() {
		var id string
		var o opRow
		if err := orows.Scan(&id, &o.action, &o.state, &o.payloadText, &o.hash); err != nil {
			orows.Close()
			fail("L.outbox", err.Error())
			return nil, global, nil
		}
		ops[strings.ToLower(id)] = o
	}
	orows.Close()
	if err := orows.Err(); err != nil {
		fail("L.outbox", err.Error())
		return nil, global, nil
	}
	var links []backlink
	brows, err := in.pool.Query(ctx,
		`SELECT event_uuid, immu_tx FROM poc_ledger_backlink
		  WHERE pg_table = 'payroll_run' AND pg_id = $1`, run)
	if err != nil {
		fail("L.backlinks", err.Error())
		return nil, global, nil
	}
	for brows.Next() {
		var l backlink
		if err := brows.Scan(&l.event, &l.tx); err != nil {
			brows.Close()
			fail("L.backlinks", err.Error())
			return nil, global, nil
		}
		l.event = strings.ToLower(l.event)
		links = append(links, l)
	}
	brows.Close()
	if err := brows.Err(); err != nil {
		fail("L.backlinks", err.Error())
		return nil, global, nil
	}
	if len(links) == 0 {
		fail("L.backlinks", "no evidence events for run")
		return nil, global, nil
	}
	for id, o := range ops {
		if (o.state == "backlinked" || o.state == "reconciled") && !hasLink(links, id) {
			fail("L.coverage", fmt.Sprintf("terminal op %s (%s) has no backlink", id, o.state))
			return nil, global, nil
		}
	}
	global = append(global, checkResult{Name: "L.coverage", Pass: true,
		Detail: fmt.Sprintf("%d events, all terminal ops linked", len(links))})

	// Per-event verification (A, B, E–J).
	calcSum := sha256.Sum256([]byte(calcVersion))
	calcDigest := hex.EncodeToString(calcSum[:])
	wantRules := map[string]string{}
	for _, p := range pins {
		wantRules[p.pack] = p.version
	}
	wantPack := ""
	if len(pins) > 0 {
		wantPack = pins[0].digest
	}
	var results []eventResult
	chain := map[string]string{}
	for _, l := range links {
		er := eventResult{EventUUID: l.event, Tx: l.tx, Verified: true}
		check := func(name string, pass bool, detail string) {
			er.Checks = append(er.Checks, checkResult{Name: name, Pass: pass, Detail: detail})
			if !pass {
				er.Verified = false
			}
		}
		o, ok := ops[l.event]
		if !ok {
			check("J.outbox-row", false, "backlink without outbox row")
			results = append(results, er)
			continue
		}
		idemKey := ledger.IdemKey(employerID, o.action, l.event)
		provKey := ledger.ProvKey(employerID, o.action, run, l.event)
		// J. Verified ledger fetch (proof-checked reads only).
		marker, origin, err := in.app.VerifiedReadMarker(ctx, idemKey)
		if err != nil {
			check("J.marker", false, err.Error())
			results = append(results, er)
			continue
		}
		got, recTx, err := in.app.VerifiedReadRecord(ctx, provKey)
		if err != nil {
			check("J.record", false, err.Error())
			results = append(results, er)
			continue
		}
		if err := in.app.VerifiedTx(ctx, l.tx); err != nil {
			check("J.tx", false, err.Error())
			results = append(results, er)
			continue
		}
		check("J.tx", true, fmt.Sprintf("tx %d attested", l.tx))
		if origin != l.tx {
			check("J.marker-origin", false, fmt.Sprintf("marker origin %d != backlink %d", origin, l.tx))
		} else {
			check("J.marker-origin", true, "")
		}
		if recTx != l.tx {
			check("J.record-origin", false, fmt.Sprintf("record origin %d != backlink %d", recTx, l.tx))
		} else {
			check("J.record-origin", true, "")
		}
		if marker != l.event+":"+got.Digest() {
			check("J.marker-bind", false, "marker does not bind event+digest")
		} else {
			check("J.marker-bind", true, "")
		}
		if got.EventUUID != l.event || got.Artifact != o.action ||
			got.PGTable != "payroll_run" || got.PGID != run || got.EmployerID != employerID {
			check("J.identity", false, fmt.Sprintf(
				"record binds artifact=%s table=%s id=%s employer=%s event=%s; want %s/payroll_run/%s/%s/%s",
				got.Artifact, got.PGTable, got.PGID, got.EmployerID, got.EventUUID,
				o.action, run, employerID, l.event))
		} else {
			check("J.identity", true, "")
		}
		// Reconstruct the expected record from PG sources.
		var sup string
		if err := in.pool.QueryRow(ctx,
			`SELECT c.op_uuid::text FROM poc_outbox c
			  WHERE c.employer_id = $1 AND c.action = 'payroll.finalized'
			    AND c.payload->>'pg_id' = $2 AND c.op_uuid <> $3::uuid
			    AND c.state IN ('backlinked', 'reconciled')
			    AND c.created_at <= (SELECT created_at FROM poc_outbox WHERE op_uuid = $3::uuid)
			  ORDER BY c.created_at DESC LIMIT 1`,
			employerID, run, l.event).Scan(&sup); err != nil {
			sup = ""
		}
		payloadSum := sha256.Sum256([]byte(o.payloadText))
		want := ledger.Record{
			EventUUID: l.event, Artifact: o.action, ArtifactID: run,
			PGTable: "payroll_run", PGID: run, EmployerID: employerID,
			SHA256:      hex.EncodeToString(payloadSum[:]),
			InputDigest: tipHash, OutputDigest: storedOut,
			RuleVersions: wantRules, PackDigest: wantPack, CalcDigest: calcDigest,
			DecidedBy: decidedBy, DecidedAt: decidedAt, Supersedes: strings.ToLower(sup),
		}
		cmp := func(name, a, b string) {
			check(name, a == b, mismatchDetail(a, b))
		}
		cmp("A.output-digest", got.OutputDigest, want.OutputDigest)
		cmp("B.input-digest", got.InputDigest, want.InputDigest)
		cmp("E.rule-pack", got.PackDigest, want.PackDigest)
		check("E.rule-versions", reflect.DeepEqual(got.RuleVersions, want.RuleVersions),
			fmt.Sprintf("ledger %v want %v", got.RuleVersions, want.RuleVersions))
		cmp("F.calc-digest", got.CalcDigest, want.CalcDigest)
		cmp("G.decided-by", got.DecidedBy, want.DecidedBy)
		cmp("G.decided-at", got.DecidedAt, want.DecidedAt)
		cmp("H.supersedes", strings.ToLower(got.Supersedes), want.Supersedes)
		cmp("I.payload-sha", got.SHA256, want.SHA256)
		// D + byte-level closure.
		check("D.canon-digest", canonDigest == got.OutputDigest,
			fmt.Sprintf("canon(persisted rows) %s vs ledger %s", canonDigest, got.OutputDigest))
		check("J.bytes", got.Digest() == want.Digest(),
			fmt.Sprintf("ledger digest %s vs reconstructed %s", got.Digest(), want.Digest()))
		chain[l.event] = strings.ToLower(got.Supersedes)
		results = append(results, er)
	}
	// H (chain): every supersedes link resolves to a known event or "".
	known := map[string]bool{}
	for _, er := range results {
		known[er.EventUUID] = true
	}
	for ev, sup := range chain {
		if sup != "" && !known[sup] {
			global = append(global, checkResult{Name: "H.chain",
				Detail: fmt.Sprintf("event %s supersedes unknown %s", ev, sup)})
			return results, global, nil
		}
	}
	global = append(global, checkResult{Name: "H.chain", Pass: true,
		Detail: fmt.Sprintf("%d links resolve", len(chain))})
	return results, global, nil
}

// parseCents converts NUMERIC(12,2) text to exact cents. No floating
// point anywhere: fractional digits beyond 2dp fail loudly (out of
// the column's scale contract).
func parseCents(s string) (int64, error) {
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("bad amount %q", s)
	}
	for _, c := range parts[0] {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("bad amount %q", s)
		}
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
		if len(frac) > 2 {
			return 0, fmt.Errorf("amount %q exceeds 2dp", s)
		}
		for _, c := range frac {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("bad amount %q", s)
			}
		}
	}
	for len(frac) < 2 {
		frac += "0"
	}
	n, err := strconv.ParseInt(parts[0]+frac, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("bad amount %q", s)
	}
	if neg {
		n = -n
	}
	return n, nil
}

// jsonEqual compares a Go value against stored JSON text semantically
// (key order and number encoding do not matter; values do).
func jsonEqual(want any, gotText string) bool {
	wantRaw, err := json.Marshal(want)
	if err != nil {
		return false
	}
	var w, g any
	if err := json.Unmarshal(wantRaw, &w); err != nil {
		return false
	}
	dec := json.NewDecoder(strings.NewReader(gotText))
	dec.UseNumber()
	var g2 any
	if err := dec.Decode(&g2); err != nil {
		return false
	}
	g = normalizeNumbers(g2)
	w = normalizeNumbers(w)
	return reflect.DeepEqual(w, g)
}

// normalizeNumbers renders every JSON number by its canonical value
// form so int64-vs-float64 encoding never compares unequal.
func normalizeNumbers(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, e := range t {
			out[k] = normalizeNumbers(e)
		}
		return out
	case []any:
		for i, e := range t {
			t[i] = normalizeNumbers(e)
		}
		return t
	case json.Number:
		return t.String()
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return v
	}
}

type backlink struct {
	event string
	tx    uint64
}

func hasLink(links []backlink, event string) bool {
	for _, l := range links {
		if l.event == event {
			return true
		}
	}
	return false
}

func mismatchDetail(a, b string) string {
	if a == b {
		return ""
	}
	return fmt.Sprintf("ledger %q want %q", trunc(a), trunc(b))
}

func trunc(s string) string {
	if len(s) > 64 {
		return s[:64] + "…"
	}
	return s
}
