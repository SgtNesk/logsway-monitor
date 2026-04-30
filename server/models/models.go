package models

import "time"

// MetricPayload — payload inviato dall'agent
type MetricPayload struct {
	Hostname  string            `json:"hostname"`
	Timestamp time.Time         `json:"timestamp"`
	Tags      []string          `json:"tags"`
	Metrics   map[string]float64 `json:"metrics"`
}

// Host — stato corrente di un host
type Host struct {
	Hostname    string            `json:"hostname"`
	Tags        []string          `json:"tags"`
	Status      string            `json:"status"` // healthy | warning | critical | offline
	LastSeen    time.Time         `json:"last_seen"`
	LastMetrics map[string]float64 `json:"last_metrics"`
}

// MetricPoint — punto storico
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
}

// DashboardStats — risposta /api/v1/stats
type DashboardStats struct {
	TotalHosts    int `json:"total_hosts"`
	HealthyHosts  int `json:"healthy_hosts"`
	WarningHosts  int `json:"warning_hosts"`
	CriticalHosts int `json:"critical_hosts"`
	OfflineHosts  int `json:"offline_hosts"`
}

// Thresholds — defaults, overridden by config file or env vars at startup
var (
	CPUWarning   = 70.0
	CPUCritical  = 85.0
	RAMWarning   = 75.0
	RAMCritical  = 90.0
	DiskWarning  = 80.0
	DiskCritical = 90.0
)

const OfflineAfter = 2 * 60 // secondi
