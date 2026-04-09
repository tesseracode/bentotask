package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(data)
}

// respondError writes a JSON error response per ADR-004.
func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// respondBadRequest writes a 400 error.
func respondBadRequest(w http.ResponseWriter, message string) {
	respondError(w, http.StatusBadRequest, "bad_request", message)
}

// respondNotFound writes a 404 error.
func respondNotFound(w http.ResponseWriter, message string) {
	respondError(w, http.StatusNotFound, "not_found", message)
}

// respondConflict writes a 409 error.
func respondConflict(w http.ResponseWriter, message string) {
	respondError(w, http.StatusConflict, "conflict", message)
}

// respondValidationError writes a 422 error.
func respondValidationError(w http.ResponseWriter, message string) {
	respondError(w, http.StatusUnprocessableEntity, "validation_error", message)
}

// respondInternalError logs the real error and returns a generic 500 response.
func respondInternalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	respondError(w, http.StatusInternalServerError, "internal_error", "an internal error occurred")
}

// collectionResponse wraps items in the ADR-004 envelope.
type collectionResponse struct {
	Items any `json:"items"`
	Count int `json:"count"`
}
