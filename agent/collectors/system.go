package collectors

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Metrics contiene tutte le metriche raccolte
type Metrics map[string]float64

// isEnabled returns true if the pointer is nil (default on) or points to true
func isEnabled(flag *bool) bool {
	return flag == nil || *flag
}

// Collect raccoglie le metriche di sistema, rispettando i flag collect
func Collect(collectCPU, collectMem, collectDisk, collectNet, collectLoad *bool) (Metrics, error) {
	m := make(Metrics)

	// CPU
	if isEnabled(collectCPU) {
		if pcts, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(pcts) > 0 {
			m["cpu_percent"] = round2(pcts[0])
		}
	}

	// RAM
	if isEnabled(collectMem) {
		if vm, err := mem.VirtualMemory(); err == nil {
			m["ram_percent"] = round2(vm.UsedPercent)
			m["ram_used_gb"] = round2(float64(vm.Used) / 1e9)
			m["ram_total_gb"] = round2(float64(vm.Total) / 1e9)
		}
	}

	// Disk (root)
	if isEnabled(collectDisk) {
		if du, err := disk.Usage("/"); err == nil {
			m["disk_percent"] = round2(du.UsedPercent)
			m["disk_used_gb"] = round2(float64(du.Used) / 1e9)
			m["disk_total_gb"] = round2(float64(du.Total) / 1e9)
		}
	}

	// Load average
	if isEnabled(collectLoad) {
		if la, err := load.Avg(); err == nil {
			m["load_1"] = round2(la.Load1)
			m["load_5"] = round2(la.Load5)
			m["load_15"] = round2(la.Load15)
		}
	}

	// Network (somma tutte le interfacce non-loopback)
	if isEnabled(collectNet) {
		if counters, err := net.IOCounters(false); err == nil && len(counters) > 0 {
			m["net_bytes_sent"] = float64(counters[0].BytesSent)
			m["net_bytes_recv"] = float64(counters[0].BytesRecv)
		}
	}

	if len(m) == 0 {
		return nil, fmt.Errorf("no metrics collected")
	}
	return m, nil
}

func round2(v float64) float64 {
	return float64(int(v*100)) / 100
}
