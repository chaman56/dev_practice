package book

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Handler turns HTTP requests into Store calls and Store results into HTTP responses.
type Handler struct {
	store  Store
	logger *slog.Logger
}

func NewHandler(store Store, logger *slog.Logger) *Handler {
	return &Handler{store: store, logger: logger}
}

// Register attaches our routes to a router.
// Since Go 1.22 the standard router understands methods ("GET") and path values ("{id}").
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /books", h.list)
	mux.HandleFunc("POST /books", h.create)
	mux.HandleFunc("GET /books/{id}", h.get)
	mux.HandleFunc("PUT /books/{id}", h.update)
	mux.HandleFunc("DELETE /books/{id}", h.delete)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	books, err := h.store.List(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	b, err := h.store.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	in, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}
	b, err := h.store.Create(r.Context(), in)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, b) // 201 = "something new was created"
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	in, ok := h.decodeAndValidate(w, r)
	if !ok {
		return
	}
	b, err := h.store.Update(r.Context(), r.PathValue("id"), in)
	if err != nil {
		h.storeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.Context(), r.PathValue("id")); err != nil {
		h.storeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent) // 204 = "done, nothing to send back"
}

// ---- helpers ----

// decodeAndValidate reads the JSON body safely and validates it.
// It writes the error response itself and returns ok=false if anything is wrong.
func (h *Handler) decodeAndValidate(w http.ResponseWriter, r *http.Request) (Input, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // reject bodies over 1 MB

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // typos like "titel" become errors instead of being ignored

	var in Input
	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return in, false
	}
	if problems := in.Validate(); len(problems) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":  "validation failed",
			"fields": problems,
		})
		return in, false
	}
	return in, true
}

// storeError maps known store errors to the right HTTP status.
func (h *Handler) storeError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	h.serverError(w, r, err)
}

// serverError logs the real error but shows the client a generic message,
// so we never leak internal details.
func (h *Handler) serverError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.Error("internal error", "err", err, "method", r.Method, "path", r.URL.Path)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
