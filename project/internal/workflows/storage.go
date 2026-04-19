package workflows

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Package-level domain errors. Storage translates Postgres failures into
// these so the handler can branch on domain concepts (not pgx specifics)
// when mapping to HTTP status codes. New domain errors should be added
// here rather than leaking raw driver errors across the boundary.
var (
	ErrNotFound     = errors.New("workflow not found")
	ErrInvalidInput = errors.New("invalid workflow input")
)

// Storage is the Postgres-backed persistence layer for workflows. It is
// a concrete type (not an interface) by design: at L01 there is exactly
// one implementation, and Go convention is to introduce interfaces at
// the consumer when a second implementation actually appears — not
// pre-declared at the producer "just in case".
type Storage struct {
	pool *pgxpool.Pool
}

// NewStorage returns a Storage backed by the given pgx connection pool.
// The pool's lifecycle is owned by the caller (typically main).
func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

// Create persists a new workflow.
func (s *Storage) Create(ctx context.Context, w *Workflow) error {
	query := `INSERT INTO workflows (name, trigger_type, steps)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`

	return s.pool.QueryRow(ctx, query, w.Name, w.TriggerType, w.Steps).
		Scan(
			&w.ID,
			&w.CreatedAt,
			&w.UpdatedAt,
		)
}

// GetByID fetches a single workflow by its id, or ErrNotFound.
func (s *Storage) GetByID(ctx context.Context, id string) (*Workflow, error) {
	query := `SELECT id, name, trigger_type, steps, created_at, updated_at from workflows
	WHERE id = $1`

	var w Workflow

	row := s.pool.QueryRow(ctx, query, id)

	err := row.Scan(&w.ID, &w.Name, &w.TriggerType, &w.Steps, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &w, nil
}

// List returns every workflow. Ordering and pagination are your call.
func (s *Storage) List(ctx context.Context) ([]*Workflow, error) {
	query := `SELECT id, name, trigger_type, steps, created_at, updated_at from workflows
	ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*Workflow{}
	for rows.Next() {
		var w Workflow
		if err := rows.Scan(&w.ID, &w.Name, &w.TriggerType, &w.Steps, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
