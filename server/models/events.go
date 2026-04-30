package models

import "time"

// Event — cambio di stato registrato per un servizio
type Event struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Hostname   string    `json:"hostname"`
	Service    string    `json:"service"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Value      *float64  `json:"value,omitempty"`
	Message    string    `json:"message"`
}

// EventsResponse — risposta /api/v1/events
type EventsResponse struct {
	Events []Event `json:"events"`
	Total  int     `json:"total"`
}

// ServiceStatus — stato di un singolo servizio
type ServiceStatus struct {
	Status    string   `json:"status"`
	Value     *float64 `json:"value,omitempty"`
	Threshold *float64 `json:"threshold,omitempty"`
}

// MatrixHost — riga nella matrice host×servizi
type MatrixHost struct {
	Hostname    string                   `json:"hostname"`
	Services    map[string]ServiceStatus `json:"services"`
	WorstStatus string                   `json:"worst_status"`
}

// MatrixResponse — risposta /api/v1/matrix
type MatrixResponse struct {
	Hosts    []MatrixHost `json:"hosts"`
	Services []string     `json:"services"`
}

// LastChange — ultimo cambio di stato
type LastChange struct {
	From string    `json:"from"`
	To   string    `json:"to"`
	At   time.Time `json:"at"`
}

// ServiceDetail — risposta /api/v1/hosts/{hostname}/services/{service}
type ServiceDetail struct {
	Hostname     string        `json:"hostname"`
	Service      string        `json:"service"`
	Status       string        `json:"status"`
	CurrentValue *float64      `json:"current_value,omitempty"`
	Thresholds   *Thresholds   `json:"thresholds,omitempty"`
	LastChange   *LastChange   `json:"last_change,omitempty"`
	RawOutput    string        `json:"raw_output"`
}

// Thresholds — soglie per un servizio
type Thresholds struct {
	Warning  float64 `json:"warning"`
	Critical float64 `json:"critical"`
}

// ServiceHistoryPoint — punto nella storia di un servizio
type ServiceHistoryPoint struct {
	Time   time.Time `json:"time"`
	Value  float64   `json:"value"`
	Status string    `json:"status"`
}

// ServiceHistoryResponse — risposta /api/v1/hosts/{h}/services/{s}/history
type ServiceHistoryResponse struct {
	Hostname string                `json:"hostname"`
	Service  string                `json:"service"`
	Points   []ServiceHistoryPoint `json:"points"`
}
