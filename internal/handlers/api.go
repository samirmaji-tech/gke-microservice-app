package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type API struct{}

func NewAPI() *API { return &API{} }

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Root -> GET / : simple liveness-friendly JSON payload.
func (a *API) Root(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "gke-microservice",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// GetItem -> GET /api/v1/items/{id} : example REST resource.
func (a *API) GetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	writeJSON(w, http.StatusOK, Item{ID: id, Name: "item-" + id})
}
