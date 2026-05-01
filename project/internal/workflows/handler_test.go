package workflows

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// fakeStore is the L02 test double for the unexported storage interface.
// Map-backed, sequential-only (no mutex), no fault injection — see ADR
// 0006 for the discipline rationale. Behavior mirrors real *Storage as
// closely as it can without driver semantics: Delete returns nil for any
// 0-rows-affected case, matching pgx UPDATE behavior. The 22P02→ErrNotFound
// translation only fires on real driver errors and is covered by the
// integration ring (see storage_integration_test.go) — not here.
type fakeStore struct {
	rows    map[string]*Workflow
	deleted map[string]bool
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		rows:    make(map[string]*Workflow),
		deleted: make(map[string]bool),
	}
}

func (f *fakeStore) Create(ctx context.Context, w *Workflow) error {
	if w.ID == "" {
		w.ID = newFakeUUID()
	}
	now := time.Now().UTC()
	if w.CreatedAt.IsZero() {
		w.CreatedAt = now
	}
	if w.UpdatedAt.IsZero() {
		w.UpdatedAt = now
	}
	f.rows[w.ID] = w
	return nil
}

func (f *fakeStore) GetByID(ctx context.Context, id string) (*Workflow, error) {
	w, ok := f.rows[id]
	if !ok || f.deleted[id] {
		return nil, ErrNotFound
	}
	return w, nil
}

func (f *fakeStore) List(ctx context.Context) ([]*Workflow, error) {
	out := []*Workflow{}
	for id, w := range f.rows {
		if f.deleted[id] {
			continue
		}
		out = append(out, w)
	}
	return out, nil
}

func (f *fakeStore) Update(ctx context.Context, id string, w *Workflow) error {
	existing, ok := f.rows[id]
	if !ok || f.deleted[id] {
		return ErrNotFound
	}
	w.ID = existing.ID
	w.CreatedAt = existing.CreatedAt
	w.UpdatedAt = time.Now().UTC()
	f.rows[id] = w
	return nil
}

func (f *fakeStore) Delete(ctx context.Context, id string) error {
	if _, ok := f.rows[id]; !ok || f.deleted[id] {
		return nil
	}
	f.deleted[id] = true
	return nil
}

func newFakeUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:4]) + "-" +
		hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" +
		hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:])
}

// errorMessage decodes the Shape 2 envelope (ADR 0004) and returns the
// message field. Migrating to Shape 3 (structured) updates this helper
// only — every assertion call site keeps working.
func errorMessage(t *testing.T, body []byte) string {
	t.Helper()
	var env struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v; body=%s", err, body)
	}
	return env.Error
}

// requireStatus fails the test if got != want.
func requireStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
}

// decodeWorkflow decodes a 2xx response body into a Workflow value.
func decodeWorkflow(t *testing.T, body []byte) Workflow {
	t.Helper()
	var w Workflow
	if err := json.Unmarshal(body, &w); err != nil {
		t.Fatalf("decode workflow: %v; body=%s", err, body)
	}
	return w
}

func newTestServer(t *testing.T, store storage) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	NewHandler(store).Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// readBody reads the entire response body and returns the bytes. Closes
// the body. Fails the test on read error.
func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return body
}

// ---- Pattern reference test ----

// TestGetByID_NotFound exercises the simplest 4xx path so the suite
// has one passing test on day one. New tests follow this shape:
// build server, fire request, assert status + envelope.
func TestGetByID_NotFound(t *testing.T) {
	srv := newTestServer(t, newFakeStore())

	resp, err := http.Get(srv.URL + "/workflows/" + newFakeUUID())
	if err != nil {
		t.Fatalf("GET: %v", err)
	}

	body := readBody(t, resp)
	requireStatus(t, resp.StatusCode, http.StatusNotFound)
	if got := errorMessage(t, body); got != "workflow not found" {
		t.Fatalf("error = %q, want %q", got, "workflow not found")
	}
}

// ---- Critical opacity invariant test (ADR 0005) ----

// TestDelete_OpacityInvariant pins ADR 0005's pairing-B opacity guarantee
// into a regression test. The contract: DELETE on each of the four ID
// classes below MUST produce byte-identical responses on the wire —
// status, body, and (modulo Date/Content-Length) headers. A future
// refactor that breaks this (e.g., adding a body to 204, leaking
// existence info via different status codes, changing envelope shape on
// one path but not another) gets caught here instead of in production.
//
// The four ID classes:
//  1. existing-not-deleted row (seed via store.Create + capture id)
//  2. just-deleted row (DELETE the same id from class 1 a second time)
//  3. never-existed UUID (well-formed UUID never inserted)
//  4. malformed UUID ("not-a-uuid") — at handler level via fakeStore,
//     this hits the same "not in rows" path as case 3. The end-to-end
//     22P02-translation behavior is covered by the integration ring.
//
// All four cases must produce status 204 and an empty body, and all
// four bodies must be bytes.Equal to each other. The byte-equality
// assertion is the strict form of "opacity-on-wire" — even if the
// expected status/body changes in a future ADR, the four cases must
// continue to match each other.
func TestDelete_OpacityInvariant(t *testing.T) {
	t.Skip("TODO(human): implement TestDelete_OpacityInvariant per the doc-comment above")
	// TODO(human): implement.
	//
	// Suggested shape (you decide the structure):
	//   1. store := newFakeStore(); seed one row via store.Create with a
	//      Workflow{Name, TriggerType: TriggerManual, Steps: []Step{}}.
	//      Capture the resulting w.ID — that's your "existing" id for case 1.
	//   2. srv := newTestServer(t, store)
	//   3. http.NewRequest("DELETE", srv.URL+"/workflows/"+id, nil)  +
	//      http.DefaultClient.Do(req) — net/http has no DELETE convenience
	//      helper like Get/Post.
	//   4. For each of the four ID classes, capture (status, body) tuples.
	//   5. Assert: every status == http.StatusNoContent (204).
	//   6. Assert: every body has len == 0  (RFC 9110 §15.3.5 — 204 no body).
	//   7. Assert: bytes.Equal across all four bodies (the opacity invariant).
	//
	// Why byte-equality and not just status checks: a regression that
	// returns 204 for three cases and 204-with-a-body for the fourth
	// passes a status-only check but leaks "this id exists" via body
	// presence. byte-equality catches that.
}
