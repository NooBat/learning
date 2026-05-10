//go:build integration

// Package integration tests for workflows.Storage. Run with:
//
//	DATABASE_URL_TEST=postgres://... go test -tags=integration ./internal/workflows/
//
// These tests exercise behavior real *Storage has that fakeStore cannot
// simulate: SQLSTATE 22P02 translation (real driver error), RETURNING
// semantics (server-generated id/timestamps), soft-delete WHERE filter
// (real SQL execution). Per ADR 0006 the ring is intentionally narrow —
// add more only when a real bug surfaces that the existing tests miss.

package workflows

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// integrationPool opens a pgxpool.Pool against DATABASE_URL_TEST and
// fails the test if the env var is unset or the connection fails. The
// pool is closed via t.Cleanup. Per-test isolation is via truncate, not
// pool-per-test — pool creation costs ~50ms and the tests share schema.
func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL_TEST")
	if dsn == "" {
		t.Skip("DATABASE_URL_TEST not set; skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// truncateWorkflows clears the workflows table so each test starts from
// a known state. TRUNCATE is cheaper than DROP/CREATE and faster than
// DELETE for small tables. Caller is responsible for calling this at
// the top of each test that mutates state.
func truncateWorkflows(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), "TRUNCATE workflows"); err != nil {
		t.Fatalf("truncate workflows: %v", err)
	}
}

// TestStorage_Update_RETURNING verifies that Update populates id,
// created_at, and updated_at from the RETURNING clause — these are
// server-managed and the handler relies on them being populated for the
// 200 OK response body. Round-1-bug class: an Update that only RETURNed
// updated_at would silently leak empty id and zero created_at into the
// response.
func TestStorage_Update_RETURNING(t *testing.T) {
	pool := integrationPool(t)
	truncateWorkflows(t, pool)
	store := NewStorage(pool)
	ctx := context.Background()

	seed := &Workflow{
		Name:        "seed",
		TriggerType: TriggerManual,
		Steps:       []Step{},
	}
	if err := store.Create(ctx, seed); err != nil {
		t.Fatalf("seed Create: %v", err)
	}

	// Postgres now() has microsecond resolution; without a deliberate
	// gap, fast hardware can produce identical timestamps for the
	// Create and Update transactions, which would defeat the .After
	// assertion below.
	time.Sleep(2 * time.Millisecond)

	update := &Workflow{
		Name:        "renamed",
		TriggerType: TriggerWebhook,
		Steps:       []Step{},
	}
	if err := store.Update(ctx, seed.ID, update); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if update.ID != seed.ID {
		t.Errorf("RETURNING id mismatch: got %q, want %q", update.ID, seed.ID)
	}
	if !update.CreatedAt.Equal(seed.CreatedAt) {
		t.Errorf("CreatedAt mutated by Update: got %v, want %v", update.CreatedAt, seed.CreatedAt)
	}
	if !update.UpdatedAt.After(seed.UpdatedAt) {
		t.Errorf("UpdatedAt not bumped by Update: got %v, was %v", update.UpdatedAt, seed.UpdatedAt)
	}
}

// TestStorage_GetByID_22P02_to_NotFound exercises ADR 0003 end-to-end:
// a malformed UUID hits SQLSTATE 22P02 from pgx, which translatePgError
// converts into ErrNotFound. fakeStore can't simulate this — only a
// real driver round-trip produces a real *pgconn.PgError with code
// 22P02. This test is the single source of truth that the translation
// rule still holds against the live driver.
func TestStorage_GetByID_22P02_to_NotFound(t *testing.T) {
	pool := integrationPool(t)
	store := NewStorage(pool)
	ctx := context.Background()

	_, err := store.GetByID(ctx, "not-a-uuid")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Malformed ID not returning correct error, got: %v, want %v", err, ErrNotFound)
	}
}

// TestStorage_Delete_SoftDeleteFilter verifies that soft-deleted rows
// are excluded from both GetByID and List — the WHERE deleted_at IS
// NULL filter must be applied uniformly across read sites. A regression
// that drops the filter on one read but not the other (e.g., a future
// query helper that forgets to compose the filter) leaks soft-deleted
// rows back to clients.
func TestStorage_Delete_SoftDeleteFilter(t *testing.T) {
	pool := integrationPool(t)
	truncateWorkflows(t, pool)
	store := NewStorage(pool)
	ctx := context.Background()

	alive := &Workflow{Name: "alive", TriggerType: TriggerManual, Steps: []Step{}}
	if err := store.Create(ctx, alive); err != nil {
		t.Fatalf("Create alive: %v", err)
	}
	tombstone := &Workflow{Name: "tombstone", TriggerType: TriggerManual, Steps: []Step{}}
	if err := store.Create(ctx, tombstone); err != nil {
		t.Fatalf("Create tombstone: %v", err)
	}

	if err := store.Delete(ctx, tombstone.ID); err != nil {
		t.Fatalf("Delete tombstone: %v", err)
	}

	if _, err := store.GetByID(ctx, tombstone.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetByID after Delete: got err=%v, want ErrNotFound", err)
	}

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("List len: got %d, want 1 (tombstone should be filtered)", len(list))
	}
	if list[0].ID != alive.ID {
		t.Errorf("List[0].ID: got %q, want %q (alive)", list[0].ID, alive.ID)
	}
}
