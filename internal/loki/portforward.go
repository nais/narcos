package loki

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

// lokiPortStr is the string literal used in exec.Command args to avoid gosec G204.
const (
	lokiAPIBase = "http://localhost:3100/loki/api/v1"
)

type portForward struct {
	cmd *exec.Cmd
}

// startPortForward launches kubectl port-forward to the Loki compactor and
// waits until the port is accepting connections (up to 10 seconds).
func startPortForward() (*portForward, error) {
	cmd := exec.Command(
		"kubectl", "port-forward",
		"-n", "nais-system",
		"loki-compactor-0",
		"3100:3100",
	)
	// Suppress kubectl output so it doesn't clutter the terminal.
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting kubectl port-forward: %w", err)
	}

	pf := &portForward{cmd: cmd}

	if err := pf.waitReady(); err != nil {
		_ = pf.stop()
		return nil, err
	}

	return pf, nil
}

// waitReady polls the Loki delete endpoint until it responds or the deadline
// is exceeded, using a 500 ms interval.
func (pf *portForward) waitReady() error {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(10 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(lokiAPIBase + "/delete")
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timed out waiting for Loki port-forward to be ready (is loki-compactor-0 running in nais-system?)")
}

// stop kills the port-forward process.
func (pf *portForward) stop() error {
	if pf.cmd != nil && pf.cmd.Process != nil {
		return pf.cmd.Process.Kill()
	}
	return nil
}
