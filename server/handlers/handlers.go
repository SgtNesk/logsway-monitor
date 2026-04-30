package handlers

import (
	"encoding/json"
	"fmt"
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
	api.HandleFunc("/ack", h.AckProblem).Methods("POST")
	api.HandleFunc("/ack", h.GetAcks).Methods("GET")
	api.HandleFunc("/ack/{hostname}/{service}", h.DeleteAck).Methods("DELETE")
	api.HandleFunc("/hosts", h.ListHosts).Methods("GET")
	api.HandleFunc("/hosts/{hostname}", h.GetHost).Methods("GET")
	api.HandleFunc("/hosts/{hostname}/metrics", h.GetHostMetrics).Methods("GET")
	api.HandleFunc("/hosts/{hostname}/services/{service}", h.GetServiceDetail).Methods("GET")
	api.HandleFunc("/hosts/{hostname}/services/{service}/history", h.GetServiceHistory).Methods("GET")
	api.HandleFunc("/stats", h.GetStats).Methods("GET")
	api.HandleFunc("/matrix", h.GetMatrix).Methods("GET")
	api.HandleFunc("/events", h.GetEvents).Methods("GET")
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
}

func (h *Handler) AckProblem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Hostname string `json:"hostname"`
		Service  string `json:"service"`
		Message  string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.Hostname == "" || req.Service == "" {
		respondError(w, http.StatusBadRequest, "hostname and service required")
		return
	}
	if req.Message == "" {
		req.Message = "Acknowledged"
	}

	if err := h.db.AckProblem(req.Hostname, req.Service, req.Message); err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	respond(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) DeleteAck(w http.ResponseWriter, r *http.Request) {
	hostname := mux.Vars(r)["hostname"]
	service := mux.Vars(r)["service"]

	if err := h.db.RemoveAck(hostname, service); err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetAcks(w http.ResponseWriter, r *http.Request) {
	acks, err := h.db.GetAllAcks()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	respond(w, http.StatusOK, acks)
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

	// Carica stato precedente PRIMA di salvare le nuove metriche
	prevHost, _ := h.db.GetHost(payload.Hostname)
	prevCustomChecks, _ := h.db.GetLatestCustomChecks(payload.Hostname)
	prevCustomByName := map[string]models.CustomCheck{}
	for _, c := range prevCustomChecks {
		prevCustomByName[c.Name] = c
	}

	if err := h.db.StoreMetrics(&payload); err != nil {
		respondError(w, http.StatusInternalServerError, "storage error")
		return
	}
	if err := h.db.SaveCustomChecks(payload.Hostname, payload.CustomChecks, payload.Timestamp); err != nil {
		respondError(w, http.StatusInternalServerError, "custom checks storage error")
		return
	}

	// Genera eventi per ogni cambio di stato servizio
	if prevHost != nil {
		now := time.Now().UTC()
		prevSvcs := serviceStatusesFromMetrics(prevHost.LastMetrics, prevHost.LastSeen)
		currSvcs := serviceStatusesFromMetrics(payload.Metrics, now)

		for _, svc := range serviceList {
			prev := prevSvcs[svc]
			curr := currSvcs[svc]
			if prev.Status == "unknown" || prev.Status == curr.Status {
				continue
			}
			h.db.CreateEvent(models.Event{ //nolint:errcheck
				Timestamp:  now,
				Hostname:   payload.Hostname,
				Service:    svc,
				FromStatus: prev.Status,
				ToStatus:   curr.Status,
				Value:      curr.Value,
				Message:    eventMessage(svc, prev.Status, curr.Status, curr.Value),
			})
		}

		for _, curr := range payload.CustomChecks {
			if curr.Name == "" {
				continue
			}
			prev, ok := prevCustomByName[curr.Name]
			if !ok || prev.Status == "" || prev.Status == curr.Status {
				continue
			}
			message := curr.Message
			if message == "" {
				message = eventMessage(curr.Name, prev.Status, curr.Status, curr.Value)
			}
			h.db.CreateEvent(models.Event{ //nolint:errcheck
				Timestamp:  now,
				Hostname:   payload.Hostname,
				Service:    curr.Name,
				FromStatus: prev.Status,
				ToStatus:   curr.Status,
				Value:      curr.Value,
				Message:    message,
			})
		}
	}

	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

func eventMessage(service, from, to string, value *float64) string {
	if to == "ok" {
		return fmt.Sprintf("%s returned to normal", service)
	}
	if value != nil {
		return fmt.Sprintf("%s is %s (value: %.1f)", service, to, *value)
	}
	return fmt.Sprintf("%s changed from %s to %s", service, from, to)
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
