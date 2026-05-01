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
	"os"
	"testing"

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
	t.Skip("TODO: implement after warm-up #3 handler tests land")
	// Suggested shape:
	//   1. pool := integrationPool(t); truncateWorkflows(t, pool)
	//   2. store := NewStorage(pool)
	//   3. seed := &Workflow{Name: "seed", TriggerType: TriggerManual, Steps: []Step{}}
	//      store.Create(ctx, seed)  → captures seed.ID, seed.CreatedAt
	//   4. update := &Workflow{Name: "renamed", TriggerType: TriggerWebhook, Steps: []Step{}}
	//      store.Update(ctx, seed.ID, update)
	//   5. Assert: update.ID == seed.ID  (RETURNING clause echoed it back)
	//   6. Assert: update.CreatedAt == seed.CreatedAt  (immutable)
	//   7. Assert: update.UpdatedAt.After(seed.UpdatedAt)  (server bumped it)
}

// TestStorage_GetByID_22P02_to_NotFound exercises ADR 0003 end-to-end:
// a malformed UUID hits SQLSTATE 22P02 from pgx, which translatePgError
// converts into ErrNotFound. fakeStore can't simulate this — only a
// real driver round-trip produces a real *pgconn.PgError with code
// 22P02. This test is the single source of truth that the translation
// rule still holds against the live driver.
func TestStorage_GetByID_22P02_to_NotFound(t *testing.T) {
	t.Skip("TODO: implement after warm-up #3 handler tests land")
	// Suggested shape:
	//   1. pool := integrationPool(t)  (no truncate needed; read-only)
	//   2. store := NewStorage(pool)
	//   3. _, err := store.GetByID(ctx, "not-a-uuid")
	//   4. Assert: errors.Is(err, ErrNotFound)
	//   5. (Optional) verify the error wraps a *pgconn.PgError with Code 22P02
	//      via errors.As, to lock the exact translation pathway. But the
	//      domain-level errors.Is assertion is the load-bearing one.
}

// TestStorage_Delete_SoftDeleteFilter verifies that soft-deleted rows
// are excluded from both GetByID and List — the WHERE deleted_at IS
// NULL filter must be applied uniformly across read sites. A regression
// that drops the filter on one read but not the other (e.g., a future
// query helper that forgets to compose the filter) leaks soft-deleted
// rows back to clients.
func TestStorage_Delete_SoftDeleteFilter(t *testing.T) {
	t.Skip("TODO: implement after warm-up #3 handler tests land")
	// Suggested shape:
	//   1. pool := integrationPool(t); truncateWorkflows(t, pool)
	//   2. store := NewStorage(pool)
	//   3. Create two workflows: alive + tombstone
	//   4. store.Delete(ctx, tombstone.ID)
	//   5. Assert: GetByID(tombstone.ID) returns ErrNotFound
	//   6. Assert: List() returns exactly [alive] — tombstone excluded
}
