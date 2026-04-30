package collectors

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CustomCheck rappresenta il risultato di uno script custom.
type CustomCheck struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Value   *float64 `json:"value,omitempty"`
	Message string   `json:"message"`
}

// RunCustomChecks esegue tutti gli script .sh eseguibili nella cartella.
func RunCustomChecks(checksDir string) []CustomCheck {
	results := []CustomCheck{}
	if checksDir == "" {
		return results
	}
	if _, err := os.Stat(checksDir); os.IsNotExist(err) {
		return results
	}

	files, err := filepath.Glob(filepath.Join(checksDir, "*.sh"))
	if err != nil {
		return results
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil || info.Mode()&0111 == 0 {
			continue
		}
		results = append(results, executeCheck(file))
	}

	return results
}

func executeCheck(scriptPath string) CustomCheck {
	baseName := filepath.Base(scriptPath)
	name := strings.TrimSuffix(baseName, ".sh")
	name = strings.TrimPrefix(name, "check_")

	check := CustomCheck{Name: name, Status: "unknown"}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		check.Status = "critical"
		check.Message = "Script execution failed: " + err.Error()
		out := strings.TrimSpace(string(output))
		if out != "" {
			check.Message += "\n" + out
		}
		return check
	}

	lines := strings.Split(strings.ReplaceAll(string(output), "\r\n", "\n"), "\n")
	if len(lines) >= 1 {
		status := strings.ToLower(strings.TrimSpace(lines[0]))
		switch status {
		case "ok", "warning", "critical":
			check.Status = status
		}
	}

	if len(lines) >= 2 {
		if val, err := strconv.ParseFloat(strings.TrimSpace(lines[1]), 64); err == nil {
			v := val
			check.Value = &v
		}
	}

	if len(lines) >= 3 {
		msg := strings.TrimSpace(strings.Join(lines[2:], "\n"))
		check.Message = msg
	}
	if check.Message == "" {
		check.Message = strings.TrimSpace(string(output))
	}

	return check
}
