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

func TestDelete_OpacityInvariant(t *testing.T) {
	store := newFakeStore()
	w := Workflow{Name: "test workflow", TriggerType: TriggerManual, Steps: []Step{}}

	err := store.Create(context.Background(), &w)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	srv := newTestServer(t, store)

	cases := []struct {
		label string
		id    string
	}{
		{"existing", w.ID},
		{"deleted", w.ID},
		{"non-existent", newFakeUUID()},
		{"malformed", "malformed-uuid"},
	}

	for _, tc := range cases {
		req, err := http.NewRequest("DELETE", srv.URL+"/workflows/"+tc.id, nil)
		if err != nil {
			t.Fatalf("[%s] new request: %v", tc.label, err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("[%s] exec request %v", tc.label, err)
		}

		requireStatus(t, res.StatusCode, http.StatusNoContent)
		if got := readBody(t, res); len(got) != 0 {
			t.Fatalf("[%s] body length = %d, want = %d", tc.label, len(got), 0)
		}
	}
}
