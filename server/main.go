package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/logsway/server/handlers"
	"github.com/logsway/server/models"
	"github.com/logsway/server/storage"
	"gopkg.in/yaml.v3"
)

// Version is set at build time via -ldflags
var Version = "dev"

// frontendFS embeds the built frontend assets.
// The build script copies frontend/dist into server/frontend/dist before building.
//
//go:embed frontend/dist
var frontendFS embed.FS

// ServerConfig mirrors the YAML config file structure.
type ServerConfig struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	Retention struct {
		Days int `yaml:"days"`
	} `yaml:"retention"`
	Thresholds struct {
		CPUWarning   float64 `yaml:"cpu_warning"`
		CPUCritical  float64 `yaml:"cpu_critical"`
		RAMWarning   float64 `yaml:"memory_warning"`
		RAMCritical  float64 `yaml:"memory_critical"`
		DiskWarning  float64 `yaml:"disk_warning"`
		DiskCritical float64 `yaml:"disk_critical"`
	} `yaml:"thresholds"`
}

func defaultConfig() *ServerConfig {
	cfg := &ServerConfig{}
	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 8080
	cfg.Database.Path = envOr("LOGSWAY_DB", "logsway.db")
	cfg.Retention.Days = 7
	cfg.Thresholds.CPUWarning = models.CPUWarning
	cfg.Thresholds.CPUCritical = models.CPUCritical
	cfg.Thresholds.RAMWarning = models.RAMWarning
	cfg.Thresholds.RAMCritical = models.RAMCritical
	cfg.Thresholds.DiskWarning = models.DiskWarning
	cfg.Thresholds.DiskCritical = models.DiskCritical
	return cfg
}

func loadConfig(path string) (*ServerConfig, error) {
	cfg := defaultConfig()
	if path == "" {
		return cfg, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config %s: %w", path, err)
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func main() {
	configPath := flag.String("config", envOr("LOGSWAY_CONFIG", ""), "Path to YAML config file")
	// Legacy flags — still work and override config file values
	addrOverride := flag.String("addr", envOr("LOGSWAY_ADDR", ""), "Override listen address (e.g. :9090)")
	dbOverride := flag.String("db", envOr("LOGSWAY_DB", ""), "Override database path")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Apply legacy flag overrides
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	if *addrOverride != "" {
		addr = *addrOverride
	}
	dbPath := cfg.Database.Path
	if *dbOverride != "" {
		dbPath = *dbOverride
	}

	// Push thresholds into models package
	models.CPUWarning = cfg.Thresholds.CPUWarning
	models.CPUCritical = cfg.Thresholds.CPUCritical
	models.RAMWarning = cfg.Thresholds.RAMWarning
	models.RAMCritical = cfg.Thresholds.RAMCritical
	models.DiskWarning = cfg.Thresholds.DiskWarning
	models.DiskCritical = cfg.Thresholds.DiskCritical

	db, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	// Background cleanup goroutine
	go func() {
		t := time.NewTicker(1 * time.Hour)
		defer t.Stop()
		for range t.C {
			db.Cleanup(time.Duration(cfg.Retention.Days) * 24 * time.Hour)
		}
	}()

	r := mux.NewRouter()
	r.Use(corsMiddleware)

	h := handlers.New(db)
	h.RegisterRoutes(r)

	// Serve embedded frontend — only if the dist directory was embedded at build time
	if sub, err := fs.Sub(frontendFS, "frontend/dist"); err == nil {
		r.PathPrefix("/").Handler(spaHandler(http.FS(sub)))
	}

	log.Printf("Logsway server v%s listening on %s (db: %s, retention: %dd)",
		Version, addr, dbPath, cfg.Retention.Days)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// spaHandler wraps a static file server to return index.html for unknown paths (SPA routing)
func spaHandler(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to open the file; on error serve index.html
		f, err := fs.Open(r.URL.Path)
		if err != nil {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
