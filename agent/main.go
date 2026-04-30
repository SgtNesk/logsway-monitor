package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/logsway/agent/collectors"
	"gopkg.in/yaml.v3"
)

// Version is set at build time via -ldflags
var Version = "dev"

type Config struct {
	Server struct {
		URL     string `yaml:"url"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"server"`
	Agent struct {
		Hostname string   `yaml:"hostname"`
		Interval int      `yaml:"interval"`
		Tags     []string `yaml:"tags"`
	} `yaml:"agent"`
	Collect struct {
		CPU     *bool `yaml:"cpu"`
		Memory  *bool `yaml:"memory"`
		Disk    *bool `yaml:"disk"`
		Network *bool `yaml:"network"`
		Load    *bool `yaml:"load"`
	} `yaml:"collect"`
}

type MetricPayload struct {
	Hostname  string             `json:"hostname"`
	Timestamp time.Time          `json:"timestamp"`
	Tags      []string           `json:"tags"`
	Metrics   collectors.Metrics `json:"metrics"`
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	hostname := cfg.Agent.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	interval := cfg.Agent.Interval
	if interval <= 0 {
		interval = 30
	}

	timeout := cfg.Server.Timeout
	if timeout <= 0 {
		timeout = 10
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	endpoint := cfg.Server.URL + "/api/v1/metrics"

	log.Printf("Logsway agent v%s starting — host=%s server=%s interval=%ds",
		Version, hostname, cfg.Server.URL, interval)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	send(client, endpoint, hostname, cfg)
	for range ticker.C {
		send(client, endpoint, hostname, cfg)
	}
}

func send(client *http.Client, endpoint, hostname string, cfg *Config) {
	m, err := collectors.Collect(cfg.Collect.CPU, cfg.Collect.Memory,
		cfg.Collect.Disk, cfg.Collect.Network, cfg.Collect.Load)
	if err != nil {
		log.Printf("[error] collect: %v", err)
		return
	}

	payload := MetricPayload{
		Hostname:  hostname,
		Timestamp: time.Now().UTC(),
		Tags:      cfg.Agent.Tags,
		Metrics:   m,
	}

	body, _ := json.Marshal(payload)
	resp, err := client.Post(endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[error] send: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[warn] server returned %d", resp.StatusCode)
		return
	}
	log.Printf("[ok] metrics sent — cpu=%.1f%% ram=%.1f%% disk=%.1f%%",
		m["cpu_percent"], m["ram_percent"], m["disk_percent"])
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	if cfg.Server.URL == "" {
		return nil, fmt.Errorf("server.url is required")
	}
	return &cfg, nil
}

