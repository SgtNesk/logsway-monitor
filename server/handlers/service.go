package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/logsway/server/models"
)

// serviceMetricMap mappa nome servizio → chiave nella metrics map
var serviceMetricMap = map[string]string{
	"cpu":     "cpu_percent",
	"memory":  "ram_percent",
	"disk":    "disk_percent",
	"load":    "load_1",
	"network": "net_bytes_sent",
}

// serviceThresholds restituisce warning e critical per un servizio
func serviceThresholds(service string) (float64, float64, bool) {
	switch service {
	case "cpu":
		return models.CPUWarning, models.CPUCritical, true
	case "memory":
		return models.RAMWarning, models.RAMCritical, true
	case "disk":
		return models.DiskWarning, models.DiskCritical, true
	case "load":
		return models.LoadWarning, models.LoadCritical, true
	}
	return 0, 0, false
}

// GetServiceDetail gestisce GET /api/v1/hosts/{hostname}/services/{service}
func (h *Handler) GetServiceDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]
	service := vars["service"]

	if !isStandardService(service) {
		check, err := h.db.GetLatestCustomCheck(hostname, service)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		if check == nil {
			respondError(w, http.StatusNotFound, "check not found")
			return
		}

		detail := models.ServiceDetail{
			Hostname:     hostname,
			Service:      service,
			Status:       check.Status,
			CurrentValue: check.Value,
			RawOutput:    check.Message,
		}

		lastEvt, _ := h.db.GetLastEvent(hostname, service)
		if lastEvt != nil {
			detail.LastChange = &models.LastChange{
				From: lastEvt.FromStatus,
				To:   lastEvt.ToStatus,
				At:   lastEvt.Timestamp,
			}
		}

		respond(w, http.StatusOK, detail)
		return
	}

	host, err := h.db.GetHost(hostname)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}
	if host == nil {
		respondError(w, http.StatusNotFound, "host not found")
		return
	}

	svcs := serviceStatusesFromMetrics(host.LastMetrics, host.LastSeen)
	svcStatus, ok := svcs[service]
	if !ok {
		respondError(w, http.StatusNotFound, "unknown service")
		return
	}

	detail := models.ServiceDetail{
		Hostname:  hostname,
		Service:   service,
		Status:    svcStatus.Status,
		RawOutput: formatRawOutput(service, host.LastMetrics, host.LastSeen),
	}

	if svcStatus.Value != nil {
		detail.CurrentValue = svcStatus.Value
	}

	if warn, crit, has := serviceThresholds(service); has {
		detail.Thresholds = &models.Thresholds{Warning: warn, Critical: crit}
	}

	// Recupera ultimo cambio di stato dal log eventi
	lastEvt, _ := h.db.GetLastEvent(hostname, service)
	if lastEvt != nil {
		detail.LastChange = &models.LastChange{
			From: lastEvt.FromStatus,
			To:   lastEvt.ToStatus,
			At:   lastEvt.Timestamp,
		}
	}

	respond(w, http.StatusOK, detail)
}

// GetServiceHistory gestisce GET /api/v1/hosts/{hostname}/services/{service}/history?hours=24
func (h *Handler) GetServiceHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]
	service := vars["service"]

	hours := 24
	if hq := r.URL.Query().Get("hours"); hq != "" {
		if v, err := strconv.Atoi(hq); err == nil && v > 0 && v <= 168 {
			hours = v
		}
	}

	if !isStandardService(service) {
		checks, err := h.db.GetCustomCheckHistory(hostname, service, time.Duration(hours)*time.Hour)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "db error")
			return
		}
		points := make([]models.ServiceHistoryPoint, 0, len(checks))
		for _, c := range checks {
			if c.Value == nil {
				continue
			}
			points = append(points, models.ServiceHistoryPoint{
				Time:   c.Timestamp,
				Value:  *c.Value,
				Status: c.Status,
			})
		}
		respond(w, http.StatusOK, models.ServiceHistoryResponse{
			Hostname: hostname,
			Service:  service,
			Points:   points,
		})
		return
	}

	metricName, ok := serviceMetricMap[service]
	if !ok {
		respondError(w, http.StatusNotFound, "unknown service")
		return
	}

	points, err := h.db.GetMetricHistory(hostname, metricName, time.Duration(hours)*time.Hour)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}

	warn, crit, _ := serviceThresholds(service)
	histPoints := make([]models.ServiceHistoryPoint, 0, len(points))
	for _, p := range points {
		status := "ok"
		if crit > 0 && p.Value >= crit {
			status = "critical"
		} else if warn > 0 && p.Value >= warn {
			status = "warning"
		}
		histPoints = append(histPoints, models.ServiceHistoryPoint{
			Time:   p.Timestamp,
			Value:  p.Value,
			Status: status,
		})
	}

	respond(w, http.StatusOK, models.ServiceHistoryResponse{
		Hostname: hostname,
		Service:  service,
		Points:   histPoints,
	})
}

func isStandardService(service string) bool {
	for _, s := range serviceList {
		if s == service {
			return true
		}
	}
	return false
}
