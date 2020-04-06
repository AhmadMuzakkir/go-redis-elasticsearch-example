package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ahmadmuzakkir/redis-elasticsearch-go-example/store"

	"github.com/go-chi/chi"
)

type Handler struct {
	logger        *log.Logger
	router        chi.Router
	viewTracker   store.ViewTracker
	viewRetriever store.ViewRetriever
}

func NewHandler(viewTracker store.ViewTracker, viewRetriever store.ViewRetriever, logger *log.Logger) *Handler {
	h := &Handler{
		viewTracker:   viewTracker,
		viewRetriever: viewRetriever,
		logger:        logger,
	}

	r := chi.NewRouter()

	r.Route("/analytics", func(r chi.Router) {
		r.Post("/", h.handleTrackView())

		r.Get("/{id}", h.handleRetrieveView())
	})

	h.router = r

	return h
}

func (h *Handler) handleTrackView() http.HandlerFunc {
	type request struct {
		ID string `json:"id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			if err == io.EOF {
				renderError(w, http.StatusBadRequest, "body is empty")
				return
			}

			renderError(w, http.StatusBadRequest, err.Error())
			return
		}

		if req.ID == "" {
			renderError(w, http.StatusBadRequest, "data.id is empty")
			return
		}

		if err := h.viewTracker.Track(r.Context(), store.ViewTrack{
			ID:        req.ID,
			Timestamp: time.Now(),
		}); err != nil {
			renderError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) handleRetrieveView() http.HandlerFunc {
	type response struct {
		ID     string            `json:"id"`
		Counts []store.ViewCount `json:"counts"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		counts, err := h.viewRetriever.Retrieve(r.Context(), id,
			store.FiveMinute, store.OneHour, store.OneDay, store.OneWeek, store.OneMonth)
		if err != nil {
			renderError(w, http.StatusInternalServerError, err.Error())
			return
		}

		var res response
		res.ID = id
		res.Counts = counts

		render(w, http.StatusOK, res)
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func render(w http.ResponseWriter, status int, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}

func renderError(w http.ResponseWriter, status int, message string) {
	render(w, status, Error{Message: message})
}
