package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/logsway/server/models"
	"github.com/logsway/server/storage"
)

type Handler struct {
	db *storage.DB
}

func New(db *storage.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", h.Health).Methods("GET")
	api.HandleFunc("/metrics", h.ReceiveMetrics).Methods("POST")
	api.HandleFunc("/hosts", h.ListHosts).Methods("GET")
	api.HandleFunc("/hosts/{hostname}", h.GetHost).Methods("GET")
	api.HandleFunc("/hosts/{hostname}/metrics", h.GetHostMetrics).Methods("GET")
	api.HandleFunc("/stats", h.GetStats).Methods("GET")
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
}

func (h *Handler) ReceiveMetrics(w http.ResponseWriter, r *http.Request) {
	var payload models.MetricPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if payload.Hostname == "" {
		respondError(w, http.StatusBadRequest, "hostname required")
		return
	}
	if err := h.db.StoreMetrics(&payload); err != nil {
		respondError(w, http.StatusInternalServerError, "storage error")
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.db.GetHosts()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if hosts == nil {
		hosts = []*models.Host{}
	}
	respond(w, http.StatusOK, hosts)
}

func (h *Handler) GetHost(w http.ResponseWriter, r *http.Request) {
	hostname := mux.Vars(r)["hostname"]
	host, err := h.db.GetHost(hostname)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if host == nil {
		respondError(w, http.StatusNotFound, "host not found")
		return
	}
	respond(w, http.StatusOK, host)
}

func (h *Handler) GetHostMetrics(w http.ResponseWriter, r *http.Request) {
	hostname := mux.Vars(r)["hostname"]

	// hours query param (default 1)
	hours := 1
	if hq := r.URL.Query().Get("hours"); hq != "" {
		if v, err := strconv.Atoi(hq); err == nil && v > 0 && v <= 168 {
			hours = v
		}
	}

	points, err := h.db.GetHostMetrics(hostname, time.Duration(hours)*time.Hour)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if points == nil {
		points = []*models.MetricPoint{}
	}
	respond(w, http.StatusOK, points)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetStats()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	respond(w, http.StatusOK, stats)
}

// helpers
func respond(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, code int, msg string) {
	respond(w, code, map[string]string{"error": msg})
}
