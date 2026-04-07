package loki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nais/naistrix"
)

// List fetches and prints all pending log deletion requests from Loki.
func List(ctx context.Context, _ *naistrix.OutputWriter) error {
	clusterCtx, err := currentContext()
	if err != nil {
		return err
	}

	fmt.Printf("Current cluster : %s\n", clusterCtx)
	fmt.Printf("Note            : There is one Loki instance per cluster.\n")
	fmt.Printf("                  Make sure you are connected to the correct cluster.\n\n")

	fmt.Println("Starting port-forward to loki-compactor-0 in nais-system...")

	pf, err := startPortForward()
	if err != nil {
		return err
	}
	defer pf.stop() //nolint:errcheck

	fmt.Println("Fetching delete requests from Loki:")

	return fetchAndPrintDeleteRequests(ctx)
}

// fetchAndPrintDeleteRequests retrieves the current delete request list from
// the Loki compactor and pretty-prints the JSON response.
func fetchAndPrintDeleteRequests(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lokiAPIBase+"/delete", nil)
	if err != nil {
		return fmt.Errorf("building Loki list request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetching Loki delete requests: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading Loki response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("delete list API returned HTTP %d: %s", resp.StatusCode, body)
	}

	// Pretty-print the JSON response.
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		// Not JSON — print as-is.
		fmt.Println(string(body))
		return nil
	}

	formatted, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(string(body))
		return nil
	}

	fmt.Println(string(formatted))
	return nil
}
