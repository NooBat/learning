// Package httpx provides transport-layer HTTP helpers for encoding
// responses and decoding requests. Scope is narrow by design: encode,
// decode, and the error envelope. Middleware, CORS, pagination, and
// request parsing belong in separate packages.
//
// Contract (per ADR 0004):
//   - Error envelope shape: {"error": "<message>"}.
//   - HTTP status codes are chosen by the caller, not carried on
//     error types (Ownership 1 — domain errors stay transport-agnostic).
package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// WriteJSON encodes body as JSON and writes it with the given status.
// Buffer-first so a marshal failure leaves the response unflushed
// rather than half-written (headers sent, body truncated).
func WriteJSON(w http.ResponseWriter, status int, body any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		// Encode failed. Preserve the Shape 2 envelope (ADR 0004) by
		// writing the canonical error body as a literal — avoids
		// recursing through WriteJSON/WriteError.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}` + "\n"))
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(buf.Bytes())
	return err
}

// WriteError writes the canonical error envelope {"error": msg} with
// the given status. Delegates to WriteJSON so the response shape and
// Content-Type stay consistent with success responses — a future
// Shape 3 migration (ADR 0004) changes only the anonymous struct here.
func WriteError(w http.ResponseWriter, status int, msg string) {
	_ = WriteJSON(w, status, struct {
		Error string `json:"error"`
	}{Error: msg})
}

func DecodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}
