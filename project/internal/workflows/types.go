// Package workflows is the domain package for user-defined automations.
//
// At L01 it holds only the domain types. The storage and handler layers
// are intentionally deferred — their boundary is the key design decision
// of this level and will be shaped collaboratively before any code is
// written.
//
// Rule of thumb for this package: Postgres knowledge lives in storage.go,
// HTTP knowledge lives in handler.go, and this file (types.go) has no
// knowledge of either — it is the shared vocabulary they both speak.
package workflows

import "time"

// Workflow is the domain type for a user-defined automation. It mirrors
// the columns of the workflows table in schema.sql but is not coupled to
// the database or to HTTP — storage translates rows ↔ Workflow, handler
// translates JSON ↔ Workflow.
type Workflow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	TriggerType string    `json:"trigger_type"`
	Steps       []Step    `json:"steps"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Step is a single action within a workflow. At L01 the shape is
// intentionally open — execution arrives at L05, so the application
// layer does not yet interpret individual steps. Treat Step as opaque
// configuration until then.
type Step struct {
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}
