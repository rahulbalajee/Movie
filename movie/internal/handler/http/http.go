package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/rahulbalajee/Movie/movie/internal/controller/movie"
)

// Handler defines the movie handler
type Handler struct {
	ctrl *movie.Controller
}

// New creates a new movie HTTP handler
func NewHandler(ctrl *movie.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

// GetMovieDetails handles GET /movie requests
func (h *Handler) GetMovieDetails(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	details, err := h.ctrl.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, movie.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("repository get error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(details); err != nil {
		log.Printf("response encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
