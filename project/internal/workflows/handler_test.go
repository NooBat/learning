package workflows

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// decodeWorkflowList decodes a 2xx response body from GET /workflows
// into a slice. Mirrors decodeWorkflow's shape.
func decodeWorkflowList(t *testing.T, body []byte) []*Workflow {
	t.Helper()
	var list []*Workflow
	if err := json.Unmarshal(body, &list); err != nil {
		t.Fatalf("decode workflow list: %v; body=%s", err, body)
	}
	return list
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

// postJSON fires a POST request with a JSON body and a Content-Type
// header. Saves three lines per Create test.
func postJSON(t *testing.T, url, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	return resp
}

// putJSON fires a PUT request with a JSON body. net/http has no Put
// convenience helper, so this wraps NewRequest + Do.
func putJSON(t *testing.T, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("PUT new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	return resp
}

// seedWorkflow inserts a minimal valid workflow into the fakeStore and
// returns the populated value (with server-assigned id + timestamps).
// Saves five lines per test that needs a pre-existing row.
func seedWorkflow(t *testing.T, store *fakeStore, name string) *Workflow {
	t.Helper()
	w := &Workflow{Name: name, TriggerType: TriggerManual, Steps: []Step{}}
	if err := store.Create(context.Background(), w); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return w
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

// ---- Create branch coverage ----

func TestCreate_Success(t *testing.T) {
	srv := newTestServer(t, newFakeStore())

	body := `{"name":"my workflow","trigger_type":"manual","steps":[]}`
	resp := postJSON(t, srv.URL+"/workflows", body)

	got := readBody(t, resp)
	requireStatus(t, resp.StatusCode, http.StatusCreated)

	w := decodeWorkflow(t, got)
	if w.Name != "my workflow" {
		t.Fatalf("name = %q, want %q", w.Name, "my workflow")
	}
	if w.TriggerType != TriggerManual {
		t.Fatalf("trigger_type = %q, want %q", w.TriggerType, TriggerManual)
	}
	if w.ID == "" {
		t.Fatalf("id = empty, want server-generated")
	}
	if w.CreatedAt.IsZero() {
		t.Fatalf("created_at = zero, want server-stamped")
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	srv := newTestServer(t, newFakeStore())

	resp := postJSON(t, srv.URL+"/workflows", `{not json`)

	body := readBody(t, resp)
	requireStatus(t, resp.StatusCode, http.StatusBadRequest)
	if got := errorMessage(t, body); got != "invalid json" {
		t.Fatalf("error = %q, want %q", got, "invalid json")
	}
}

func TestCreate_ValidationFailures(t *testing.T) {
	cases := []struct {
		label    string
		body     string
		contains string
	}{
		{
			label:    "empty name",
			body:     `{"name":"","trigger_type":"manual","steps":[]}`,
			contains: "name",
		},
		{
			label:    "name too long",
			body:     `{"name":"` + strings.Repeat("a", 501) + `","trigger_type":"manual","steps":[]}`,
			contains: "500 characters",
		},
		{
			label:    "unknown trigger type",
			body:     `{"name":"x","trigger_type":"alien","steps":[]}`,
			contains: "trigger",
		},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			srv := newTestServer(t, newFakeStore())

			resp := postJSON(t, srv.URL+"/workflows", tc.body)
			body := readBody(t, resp)

			requireStatus(t, resp.StatusCode, http.StatusBadRequest)
			if msg := errorMessage(t, body); !strings.Contains(msg, tc.contains) {
				t.Fatalf("error = %q, want contains %q", msg, tc.contains)
			}
		})
	}
}

func TestCreate_BodyTooLarge(t *testing.T) {
	srv := newTestServer(t, newFakeStore())

	// 2 MiB > 1 MiB cap. handler.create wraps r.Body in MaxBytesReader,
	// so DecodeJSON returns *http.MaxBytesError, which the handler maps
	// to 413 with "request body too large" message.
	huge := strings.Repeat("a", 2*MiB)
	body := `{"name":"` + huge + `","trigger_type":"manual","steps":[]}`

	resp := postJSON(t, srv.URL+"/workflows", body)
	got := readBody(t, resp)

	requireStatus(t, resp.StatusCode, http.StatusRequestEntityTooLarge)
	if msg := errorMessage(t, got); msg != "request body too large" {
		t.Fatalf("error = %q, want %q", msg, "request body too large")
	}
}

// ---- Update branch coverage ----

func TestUpdate_Success(t *testing.T) {
	store := newFakeStore()
	seed := seedWorkflow(t, store, "original")
	srv := newTestServer(t, store)

	body := `{"name":"renamed","trigger_type":"webhook","steps":[]}`
	resp := putJSON(t, srv.URL+"/workflows/"+seed.ID, body)

	got := readBody(t, resp)
	requireStatus(t, resp.StatusCode, http.StatusOK)

	w := decodeWorkflow(t, got)
	if w.Name != "renamed" {
		t.Fatalf("name = %q, want %q", w.Name, "renamed")
	}
	if w.TriggerType != TriggerWebhook {
		t.Fatalf("trigger_type = %q, want %q", w.TriggerType, TriggerWebhook)
	}
	if w.ID != seed.ID {
		t.Fatalf("id = %q, want %q (id is immutable)", w.ID, seed.ID)
	}
	if !w.CreatedAt.Equal(seed.CreatedAt) {
		t.Fatalf("created_at = %v, want %v (created_at is immutable)", w.CreatedAt, seed.CreatedAt)
	}
}

func TestUpdate_InvalidJSON(t *testing.T) {
	store := newFakeStore()
	seed := seedWorkflow(t, store, "x")
	srv := newTestServer(t, store)

	resp := putJSON(t, srv.URL+"/workflows/"+seed.ID, `{not json`)
	body := readBody(t, resp)

	requireStatus(t, resp.StatusCode, http.StatusBadRequest)
	if got := errorMessage(t, body); got != "invalid json" {
		t.Fatalf("error = %q, want %q", got, "invalid json")
	}
}

// TestUpdate_NotFound covers two ID classes that must collapse to one
// response per ADR 0005's opacity stance: never-existed and soft-deleted.
// Per-ADR-0003, malformed UUID also collapses here at the integration
// ring level — at handler-level via fakeStore, malformed hits the same
// not-in-rows path as never-existed.
func TestUpdate_NotFound(t *testing.T) {
	cases := []struct {
		label string
		setup func(*fakeStore) string
	}{
		{
			label: "never existed",
			setup: func(_ *fakeStore) string {
				return newFakeUUID()
			},
		},
		{
			label: "soft deleted",
			setup: func(s *fakeStore) string {
				w := seedWorkflow(t, s, "doomed")
				if err := s.Delete(context.Background(), w.ID); err != nil {
					t.Fatalf("delete seed: %v", err)
				}
				return w.ID
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			store := newFakeStore()
			id := tc.setup(store)
			srv := newTestServer(t, store)

			body := `{"name":"x","trigger_type":"manual","steps":[]}`
			resp := putJSON(t, srv.URL+"/workflows/"+id, body)
			got := readBody(t, resp)

			requireStatus(t, resp.StatusCode, http.StatusNotFound)
			if msg := errorMessage(t, got); msg != "workflow not found" {
				t.Fatalf("[%s] error = %q, want %q", tc.label, msg, "workflow not found")
			}
		})
	}
}

func TestUpdate_BodyTooLarge(t *testing.T) {
	store := newFakeStore()
	seed := seedWorkflow(t, store, "x")
	srv := newTestServer(t, store)

	huge := strings.Repeat("a", 2*MiB)
	body := `{"name":"` + huge + `","trigger_type":"manual","steps":[]}`

	resp := putJSON(t, srv.URL+"/workflows/"+seed.ID, body)
	got := readBody(t, resp)

	requireStatus(t, resp.StatusCode, http.StatusRequestEntityTooLarge)
	if msg := errorMessage(t, got); msg != "request body too large" {
		t.Fatalf("error = %q, want %q", msg, "request body too large")
	}
}

func TestUpdate_ValidationFailures(t *testing.T) {
	cases := []struct {
		label    string
		body     string
		contains string
	}{
		{"empty name", `{"name":"","trigger_type":"manual","steps":[]}`, "name"},
		{"unknown trigger", `{"name":"x","trigger_type":"alien","steps":[]}`, "trigger"},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			store := newFakeStore()
			seed := seedWorkflow(t, store, "x")
			srv := newTestServer(t, store)

			resp := putJSON(t, srv.URL+"/workflows/"+seed.ID, tc.body)
			body := readBody(t, resp)

			requireStatus(t, resp.StatusCode, http.StatusBadRequest)
			if msg := errorMessage(t, body); !strings.Contains(msg, tc.contains) {
				t.Fatalf("error = %q, want contains %q", msg, tc.contains)
			}
		})
	}
}

// ---- List + GetByID success coverage ----

func TestList_Empty(t *testing.T) {
	srv := newTestServer(t, newFakeStore())

	resp, err := http.Get(srv.URL + "/workflows")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	body := readBody(t, resp)

	requireStatus(t, resp.StatusCode, http.StatusOK)
	if got := decodeWorkflowList(t, body); len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

// TestList_ReturnsLiveOnly exercises the soft-delete WHERE filter at
// the list endpoint. Tombstoned rows MUST NOT leak into the list
// response — same opacity principle as ADR 0005, applied to a read
// site instead of a write site.
func TestList_ReturnsLiveOnly(t *testing.T) {
	store := newFakeStore()
	alive := seedWorkflow(t, store, "alive")
	tombstone := seedWorkflow(t, store, "tombstone")
	if err := store.Delete(context.Background(), tombstone.ID); err != nil {
		t.Fatalf("delete tombstone: %v", err)
	}
	srv := newTestServer(t, store)

	resp, err := http.Get(srv.URL + "/workflows")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	body := readBody(t, resp)

	requireStatus(t, resp.StatusCode, http.StatusOK)
	got := decodeWorkflowList(t, body)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1 (tombstone excluded)", len(got))
	}
	if got[0].ID != alive.ID {
		t.Fatalf("returned id = %q, want %q (alive only)", got[0].ID, alive.ID)
	}
}

func TestGetByID_Success(t *testing.T) {
	store := newFakeStore()
	seed := seedWorkflow(t, store, "found")
	srv := newTestServer(t, store)

	resp, err := http.Get(srv.URL + "/workflows/" + seed.ID)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	body := readBody(t, resp)

	requireStatus(t, resp.StatusCode, http.StatusOK)
	w := decodeWorkflow(t, body)
	if w.ID != seed.ID {
		t.Fatalf("id = %q, want %q", w.ID, seed.ID)
	}
	if w.Name != "found" {
		t.Fatalf("name = %q, want %q", w.Name, "found")
	}
}

// TestGetByID_SoftDeleted is the read-site mirror of TestUpdate_NotFound's
// soft-deleted case. Tombstoned rows must return the same 404 response
// as never-existed rows — the opacity invariant applied to GetByID.
func TestGetByID_SoftDeleted(t *testing.T) {
	store := newFakeStore()
	seed := seedWorkflow(t, store, "doomed")
	if err := store.Delete(context.Background(), seed.ID); err != nil {
		t.Fatalf("delete seed: %v", err)
	}
	srv := newTestServer(t, store)

	resp, err := http.Get(srv.URL + "/workflows/" + seed.ID)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	body := readBody(t, resp)

	requireStatus(t, resp.StatusCode, http.StatusNotFound)
	if msg := errorMessage(t, body); msg != "workflow not found" {
		t.Fatalf("error = %q, want %q", msg, "workflow not found")
	}
}
