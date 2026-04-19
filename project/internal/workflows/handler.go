package workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"
)

// Handler is the HTTP adapter for the workflows domain. It owns JSON
// encoding/decoding, HTTP status code selection, and request routing —
// but holds no Postgres knowledge. All persistence goes through Storage.
type Handler struct {
	store storage
}

// storage is the set of storage operations Handler needs — a consumer-
// defined interface (unexported, declared here in the handler's file)
// rather than an exported interface from the producer side. Any type
// with these methods can back the Handler.
type storage interface {
	Create(ctx context.Context, w *Workflow) error
	GetByID(ctx context.Context, id string) (*Workflow, error)
	List(ctx context.Context) ([]*Workflow, error)
}

// NewHandler wires a Handler to anything that satisfies storage.
func NewHandler(store storage) *Handler {
	return &Handler{store}
}

// Register attaches the workflows routes to the given mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /workflows", h.create)
	mux.HandleFunc("GET /workflows/{id}", h.getByID)
	mux.HandleFunc("GET /workflows", h.list)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var workflow Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		log.Printf("workflows: json decoding error: %v", err)
		return
	}

	if err := validate(&workflow); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.store.Create(r.Context(), &workflow)
	if errors.Is(err, ErrInvalidInput) {
		http.Error(w, "invalid workflow input", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("workflows: unable to create: %v", err)
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(workflow); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("workflows: json encoding error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(buf.Bytes())
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	workflow, err := h.store.GetByID(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "workflow not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("workflows: unable to find %s: %v", id, err)
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(workflow); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("workflows: json encoding error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	slice, err := h.store.List(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("workflows: unable to list: %v", err)
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(slice); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("workflows: json encoding error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

// validate enforces domain invariants on an inbound Workflow before it
// reaches Storage. Returning an error wrapping ErrInvalidInput lets the
// handler's error-mapping layer turn it into a 400 with a useful
// message, without handlers needing to know the specifics.
func validate(w *Workflow) error {
	if w != nil {
		w.Name = strings.TrimSpace(w.Name)
		if w.Name == "" {
			return wrapInvalid("name must not be empty")
		}

		if utf8.RuneCountInString(w.Name) > 500 {
			return wrapInvalid("name should be less than 500 characters")
		}

		if !slices.Contains(ValidTriggerTypes, w.TriggerType) {
			return wrapInvalid("invalid trigger type: %s", w.TriggerType)
		}
	}
	return nil
}

// wrapInvalid builds an ErrInvalidInput error with a human-readable
// reason while preserving errors.Is semantics.
func wrapInvalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(format, args...))
}

// Ensure the imports stay referenced while bodies are TODO stubs.
var _ = errors.Is
var _ = json.NewDecoder
