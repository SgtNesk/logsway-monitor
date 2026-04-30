package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/logsway/server/models"
	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
}

func New(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Performance pragmas
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	db.Exec("PRAGMA foreign_keys=ON")

	s := &DB{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *DB) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS hosts (
			hostname     TEXT PRIMARY KEY,
			tags         TEXT NOT NULL DEFAULT '[]',
			status       TEXT NOT NULL DEFAULT 'offline',
			last_seen    DATETIME NOT NULL,
			last_metrics TEXT NOT NULL DEFAULT '{}'
		)`,
		`CREATE TABLE IF NOT EXISTS metrics (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			hostname  TEXT NOT NULL,
			name      TEXT NOT NULL,
			value     REAL NOT NULL,
			timestamp DATETIME NOT NULL,
			FOREIGN KEY (hostname) REFERENCES hosts(hostname) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_hostname_ts ON metrics(hostname, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_ts ON metrics(timestamp)`,
		`CREATE TABLE IF NOT EXISTS events (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp   DATETIME NOT NULL,
			hostname    TEXT NOT NULL,
			service     TEXT NOT NULL,
			from_status TEXT NOT NULL,
			to_status   TEXT NOT NULL,
			value       REAL,
			message     TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_events_hostname ON events(hostname, timestamp DESC)`,
	}
	for _, s2 := range stmts {
		if _, err := s.db.Exec(s2); err != nil {
			return err
		}
	}
	return nil
}

// StoreMetrics salva le metriche di un agent e aggiorna lo stato host
func (s *DB) StoreMetrics(payload *models.MetricPayload) error {
	status := computeStatus(payload.Metrics)

	tagsJSON, _ := json.Marshal(payload.Tags)
	metricsJSON, _ := json.Marshal(payload.Metrics)

	ts := payload.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	_, err := s.db.Exec(`
		INSERT INTO hosts (hostname, tags, status, last_seen, last_metrics)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(hostname) DO UPDATE SET
			tags         = excluded.tags,
			status       = excluded.status,
			last_seen    = excluded.last_seen,
			last_metrics = excluded.last_metrics
	`, payload.Hostname, string(tagsJSON), status, ts.UTC(), string(metricsJSON))
	if err != nil {
		return fmt.Errorf("upsert host: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO metrics (hostname, name, value, timestamp) VALUES (?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for name, value := range payload.Metrics {
		if _, err := stmt.Exec(payload.Hostname, name, value, ts.UTC()); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetHosts restituisce tutti gli host aggiornando lo stato offline
func (s *DB) GetHosts() ([]*models.Host, error) {
	rows, err := s.db.Query(`SELECT hostname, tags, status, last_seen, last_metrics FROM hosts ORDER BY hostname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []*models.Host
	for rows.Next() {
		h := &models.Host{}
		var tagsJSON, metricsJSON string
		var lastSeen time.Time
		if err := rows.Scan(&h.Hostname, &tagsJSON, &h.Status, &lastSeen, &metricsJSON); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsJSON), &h.Tags)
		json.Unmarshal([]byte(metricsJSON), &h.LastMetrics)
		h.LastSeen = lastSeen

		// Mark offline if not seen recently
		if time.Since(lastSeen).Seconds() > models.OfflineAfter {
			h.Status = "offline"
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

// GetHost restituisce un singolo host
func (s *DB) GetHost(hostname string) (*models.Host, error) {
	h := &models.Host{}
	var tagsJSON, metricsJSON string
	var lastSeen time.Time

	err := s.db.QueryRow(`SELECT hostname, tags, status, last_seen, last_metrics FROM hosts WHERE hostname = ?`, hostname).
		Scan(&h.Hostname, &tagsJSON, &h.Status, &lastSeen, &metricsJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(tagsJSON), &h.Tags)
	json.Unmarshal([]byte(metricsJSON), &h.LastMetrics)
	h.LastSeen = lastSeen

	if time.Since(lastSeen).Seconds() > models.OfflineAfter {
		h.Status = "offline"
	}
	return h, nil
}

// GetHostMetrics restituisce lo storico metriche di un host
func (s *DB) GetHostMetrics(hostname string, since time.Duration) ([]*models.MetricPoint, error) {
	rows, err := s.db.Query(`
		SELECT name, value, timestamp FROM metrics
		WHERE hostname = ? AND timestamp > ?
		ORDER BY timestamp ASC
	`, hostname, time.Now().Add(-since).UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []*models.MetricPoint
	for rows.Next() {
		p := &models.MetricPoint{}
		if err := rows.Scan(&p.Name, &p.Value, &p.Timestamp); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

// GetStats restituisce le statistiche aggregate
func (s *DB) GetStats() (*models.DashboardStats, error) {
	hosts, err := s.GetHosts()
	if err != nil {
		return nil, err
	}
	stats := &models.DashboardStats{TotalHosts: len(hosts)}
	for _, h := range hosts {
		switch h.Status {
		case "healthy":
			stats.HealthyHosts++
		case "warning":
			stats.WarningHosts++
		case "critical":
			stats.CriticalHosts++
		case "offline":
			stats.OfflineHosts++
		}
	}
	return stats, nil
}

// Cleanup elimina metriche più vecchie di retention
func (s *DB) Cleanup(retention time.Duration) {
	cutoff := time.Now().Add(-retention).UTC()
	res, err := s.db.Exec(`DELETE FROM metrics WHERE timestamp < ?`, cutoff)
	if err != nil {
		log.Printf("[cleanup] error: %v", err)
		return
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		log.Printf("[cleanup] deleted %d old metric points", n)
		s.db.Exec("PRAGMA wal_checkpoint(PASSIVE)")
	}
}

// computeStatus calcola lo stato da metriche
func computeStatus(metrics map[string]float64) string {
	status := "healthy"

	check := func(val, warn, crit float64) {
		if val >= crit {
			status = "critical"
		} else if val >= warn && status == "healthy" {
			status = "warning"
		}
	}

	if v, ok := metrics["cpu_percent"]; ok {
		check(v, models.CPUWarning, models.CPUCritical)
	}
	if v, ok := metrics["ram_percent"]; ok {
		check(v, models.RAMWarning, models.RAMCritical)
	}
	if v, ok := metrics["disk_percent"]; ok {
		check(v, models.DiskWarning, models.DiskCritical)
	}
	return status
}

// CreateEvent registra un evento di cambio stato
func (s *DB) CreateEvent(e models.Event) error {
	_, err := s.db.Exec(
		`INSERT INTO events (timestamp, hostname, service, from_status, to_status, value, message)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.Timestamp, e.Hostname, e.Service, e.FromStatus, e.ToStatus, e.Value, e.Message,
	)
	return err
}

// GetEvents restituisce eventi recenti, opzionalmente filtrati per hostname
func (s *DB) GetEvents(hours int, hostname string) ([]models.Event, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour).UTC()

	var rows *sql.Rows
	var err error
	if hostname != "" {
		rows, err = s.db.Query(
			`SELECT id, timestamp, hostname, service, from_status, to_status, value, message
			 FROM events WHERE timestamp > ? AND hostname = ? ORDER BY timestamp DESC LIMIT 500`,
			since, hostname,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT id, timestamp, hostname, service, from_status, to_status, value, message
			 FROM events WHERE timestamp > ? ORDER BY timestamp DESC LIMIT 500`,
			since,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		var val sql.NullFloat64
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.Hostname, &e.Service,
			&e.FromStatus, &e.ToStatus, &val, &e.Message); err != nil {
			return nil, err
		}
		if val.Valid {
			v := val.Float64
			e.Value = &v
		}
		events = append(events, e)
	}
	if events == nil {
		events = []models.Event{}
	}
	return events, nil
}

// GetLastEvent restituisce l'ultimo evento per host+service
func (s *DB) GetLastEvent(hostname, service string) (*models.Event, error) {
	var e models.Event
	var val sql.NullFloat64
	err := s.db.QueryRow(
		`SELECT id, timestamp, hostname, service, from_status, to_status, value, message
		 FROM events WHERE hostname = ? AND service = ? ORDER BY timestamp DESC LIMIT 1`,
		hostname, service,
	).Scan(&e.ID, &e.Timestamp, &e.Hostname, &e.Service,
		&e.FromStatus, &e.ToStatus, &val, &e.Message)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if val.Valid {
		v := val.Float64
		e.Value = &v
	}
	return &e, nil
}

// GetMetricHistory restituisce lo storico di una singola metrica
func (s *DB) GetMetricHistory(hostname, metricName string, since time.Duration) ([]*models.MetricPoint, error) {
	rows, err := s.db.Query(
		`SELECT name, value, timestamp FROM metrics
		 WHERE hostname = ? AND name = ? AND timestamp > ?
		 ORDER BY timestamp ASC`,
		hostname, metricName, time.Now().Add(-since).UTC(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []*models.MetricPoint
	for rows.Next() {
		p := &models.MetricPoint{}
		if err := rows.Scan(&p.Name, &p.Value, &p.Timestamp); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}
