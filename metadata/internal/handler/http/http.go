package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/rahulbalajee/Movie/metadata/internal/controller/metadata"
	"github.com/rahulbalajee/Movie/metadata/internal/repository"
)

// Handler defines the movie metadata HTTP handler
type Handler struct {
	ctrl *metadata.Controller
}

// Factory function to create a new movie metadata HTTP handler
func NewHandler(ctrl *metadata.Controller) *Handler {
	return &Handler{
		ctrl: ctrl,
	}
}

// GetMetadata handles GET /metadata requests  n
func (h *Handler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	m, err := h.ctrl.Get(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("repository get error for movie %s: %v\n", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(m); err != nil {
		log.Printf("response encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
