package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/rahulbalajee/Movie/rating/internal/controller/rating"
	"github.com/rahulbalajee/Movie/rating/pkg/model"
)

// Handler defines a rating service HTTP handler
type Handler struct {
	ctrl *rating.Controller
}

// Factory function to create a new rating service HTTP handler
func NewHandler(ctrl *rating.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

// Handles /GET rating requests HTTP
func (h *Handler) GetRating(w http.ResponseWriter, r *http.Request) {
	recordId := model.RecordID(r.FormValue("id"))
	if recordId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recordType := model.RecordType(r.FormValue("type"))
	if recordType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	v, err := h.ctrl.GetAggregatedRatings(r.Context(), recordId, recordType)
	if err != nil {
		if errors.Is(err, rating.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Printf("repository get error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("response encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Handles /PUT rating requests HTTP
func (h *Handler) PutRating(w http.ResponseWriter, r *http.Request) {
	recordId := model.RecordID(r.FormValue("id"))
	if recordId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recordType := model.RecordType(r.FormValue("type"))
	if recordType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := model.UserID(r.FormValue("userId"))
	if userId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	v, err := strconv.ParseFloat(r.FormValue("value"), 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = h.ctrl.PutRating(r.Context(), recordId, recordType, &model.Rating{UserID: userId, Value: model.RatingValue(v)}); err != nil {
		log.Printf("repository put error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
