package handlers

import (
	"net/http"
	"strconv"
)

// GetEvents gestisce GET /api/v1/events?hours=24&hostname=...
func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if hq := r.URL.Query().Get("hours"); hq != "" {
		if v, err := strconv.Atoi(hq); err == nil && v > 0 && v <= 720 {
			hours = v
		}
	}
	hostname := r.URL.Query().Get("hostname")

	events, err := h.db.GetEvents(hours, hostname)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}
