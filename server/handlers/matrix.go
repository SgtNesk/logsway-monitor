package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/logsway/server/models"
)

var serviceList = []string{"cpu", "memory", "disk", "load", "network", "ping"}

// GetMatrix gestisce GET /api/v1/matrix
func (h *Handler) GetMatrix(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.db.GetHosts()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "db error")
		return
	}

	result := make([]models.MatrixHost, 0, len(hosts))
	for _, host := range hosts {
		svcs := serviceStatusesFromMetrics(host.LastMetrics, host.LastSeen)
		result = append(result, models.MatrixHost{
			Hostname:    host.Hostname,
			Services:    svcs,
			WorstStatus: worstStatus(svcs),
		})
	}

	resp := models.MatrixResponse{
		Hosts:    result,
		Services: serviceList,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// serviceStatusesFromMetrics calcola lo stato di ogni servizio da una mappa di metriche
func serviceStatusesFromMetrics(metrics map[string]float64, lastSeen time.Time) map[string]models.ServiceStatus {
	svcs := make(map[string]models.ServiceStatus, len(serviceList))

	calcSvc := func(metricKey string, warn, crit float64) models.ServiceStatus {
		v, ok := metrics[metricKey]
		if !ok {
			return models.ServiceStatus{Status: "unknown"}
		}
		val := v
		threshold := warn
		status := "ok"
		if v >= crit {
			status = "critical"
			threshold = crit
		} else if v >= warn {
			status = "warning"
		}
		return models.ServiceStatus{Status: status, Value: &val, Threshold: &threshold}
	}

	svcs["cpu"] = calcSvc("cpu_percent", models.CPUWarning, models.CPUCritical)
	svcs["memory"] = calcSvc("ram_percent", models.RAMWarning, models.RAMCritical)
	svcs["disk"] = calcSvc("disk_percent", models.DiskWarning, models.DiskCritical)
	svcs["load"] = calcSvc("load_1", models.LoadWarning, models.LoadCritical)

	if _, ok := metrics["net_bytes_sent"]; ok {
		svcs["network"] = models.ServiceStatus{Status: "ok"}
	} else {
		svcs["network"] = models.ServiceStatus{Status: "unknown"}
	}

	if time.Since(lastSeen).Seconds() > float64(models.OfflineAfter) {
		svcs["ping"] = models.ServiceStatus{Status: "critical"}
	} else {
		svcs["ping"] = models.ServiceStatus{Status: "ok"}
	}

	return svcs
}

func worstStatus(svcs map[string]models.ServiceStatus) string {
	order := map[string]int{"critical": 0, "warning": 1, "ok": 2, "unknown": 3}
	worst := "unknown"
	for _, s := range svcs {
		if order[s.Status] < order[worst] {
			worst = s.Status
		}
	}
	return worst
}

// formatRawOutput formatta le metriche come testo leggibile stile Xymon
func formatRawOutput(service string, metrics map[string]float64, lastSeen time.Time) string {
	var b strings.Builder

	switch service {
	case "cpu":
		if v, ok := metrics["cpu_percent"]; ok {
			fmt.Fprintf(&b, "CPU Usage:  %.1f%%\n", v)
		}
	case "memory":
		if v, ok := metrics["ram_percent"]; ok {
			fmt.Fprintf(&b, "RAM Usage:  %.1f%%\n", v)
		}
		if v, ok := metrics["ram_used_gb"]; ok {
			fmt.Fprintf(&b, "RAM Used:   %.2f GB\n", v)
		}
		if v, ok := metrics["ram_total_gb"]; ok {
			fmt.Fprintf(&b, "RAM Total:  %.2f GB\n", v)
		}
	case "disk":
		if v, ok := metrics["disk_percent"]; ok {
			fmt.Fprintf(&b, "Disk Usage: %.1f%%\n", v)
		}
		if v, ok := metrics["disk_used_gb"]; ok {
			fmt.Fprintf(&b, "Disk Used:  %.2f GB\n", v)
		}
		if v, ok := metrics["disk_total_gb"]; ok {
			fmt.Fprintf(&b, "Disk Total: %.2f GB\n", v)
		}
	case "load":
		if v, ok := metrics["load_1"]; ok {
			fmt.Fprintf(&b, "Load  1m:  %.2f\n", v)
		}
		if v, ok := metrics["load_5"]; ok {
			fmt.Fprintf(&b, "Load  5m:  %.2f\n", v)
		}
		if v, ok := metrics["load_15"]; ok {
			fmt.Fprintf(&b, "Load 15m:  %.2f\n", v)
		}
	case "network":
		if v, ok := metrics["net_bytes_sent"]; ok {
			fmt.Fprintf(&b, "Bytes Sent: %.0f\n", v)
		}
		if v, ok := metrics["net_bytes_recv"]; ok {
			fmt.Fprintf(&b, "Bytes Recv: %.0f\n", v)
		}
	case "ping":
		age := time.Since(lastSeen)
		fmt.Fprintf(&b, "Last seen:  %s ago\n", fmtDuration(age))
		if age.Seconds() > float64(models.OfflineAfter) {
			fmt.Fprintf(&b, "Status:     OFFLINE (no data for %s)\n", fmtDuration(age))
		} else {
			fmt.Fprintf(&b, "Status:     ONLINE\n")
		}
	}

	if b.Len() == 0 {
		b.WriteString("No data available\n")
	}
	return b.String()
}

func fmtDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	return fmt.Sprintf("%.0fh", d.Hours())
}
