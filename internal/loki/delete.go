package loki

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/loki/command/flag"
)

// Delete submits a log deletion request to Loki for the given application.
func Delete(ctx context.Context, flags *flag.Delete, out *naistrix.OutputWriter) error {
	if flags.Namespace == "" {
		return fmt.Errorf("--namespace is required")
	}
	if flags.App == "" {
		return fmt.Errorf("--app is required")
	}
	if flags.Days <= 0 {
		return fmt.Errorf("--days must be a positive integer")
	}

	clusterCtx, err := currentContext()
	if err != nil {
		return err
	}

	fmt.Printf("Current cluster : %s\n", clusterCtx)
	fmt.Printf("Note            : There is one Loki instance per cluster.\n")
	fmt.Printf("                  Make sure you are connected to the correct cluster before proceeding.\n\n")

	// Calculate the start timestamp (N days ago).
	startTS := time.Now().AddDate(0, 0, -flags.Days).Unix()

	// Build the LogQL query.
	query := fmt.Sprintf(`{service_namespace=%q, service_name=%q}`, flags.Namespace, flags.App)
	if flags.Filter != "" {
		query += " | " + flags.Filter
	}
	if flags.Regex != "" {
		query += fmt.Sprintf(` |~ "(?i)%s"`, flags.Regex)
	}

	encodedQuery := url.QueryEscape(query)
	deleteURL := fmt.Sprintf("%s/delete?query=%s&start=%d", lokiAPIBase, encodedQuery, startTS)

	fmt.Printf("Deletion query  : %s\n", query)
	fmt.Printf("Start timestamp : %d (%s)\n", startTS, time.Unix(startTS, 0).Format(time.RFC3339))
	fmt.Printf("cURL equivalent : curl -g -X POST %q\n\n", deleteURL)

	ok, err := out.Confirm("Proceed with deletion?")
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Aborted.")
		return nil
	}

	fmt.Println("\nStarting port-forward to loki-compactor-0 in nais-system...")

	pf, err := startPortForward()
	if err != nil {
		return err
	}
	defer pf.stop() //nolint:errcheck

	// POST the deletion request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("building Loki delete request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling Loki delete API: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("delete API returned HTTP %d: %s", resp.StatusCode, body)
	}

	fmt.Printf("\nDeletion request submitted (HTTP %d).\n\n", resp.StatusCode)
	fmt.Println("Fetching updated list of delete requests from Loki:")

	return fetchAndPrintDeleteRequests(ctx)
}
