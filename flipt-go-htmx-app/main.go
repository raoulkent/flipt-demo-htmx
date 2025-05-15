package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	defaultFliptURL = "http://flipt:8080"
	flagKey         = "my_awesome_feature" // Replace with your flag key
	entityID        = "user123"
)

// Evaluate the feature flag, supporting both Boolean and variant flags
func evaluateFeature(fliptURL string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/evaluate", fliptURL)
	payload := map[string]interface{}{
		"flagKey":  flagKey,
		"entityId": entityID,
		"context":  map[string]interface{}{},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "error", fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	body := bytes.NewBuffer(payloadBytes)

	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return "error", fmt.Errorf("failed to make request to Flipt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "error", fmt.Errorf("Flipt returned non-OK status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "error", fmt.Errorf("failed to decode Flipt response: %w", err)
	}

	// Prefer 'enabled' (Boolean flag), else use 'value' (variant flag)
	if enabled, ok := result["enabled"].(bool); ok {
		if enabled {
			return "true", nil
		}
		return "false", nil
	}
	if value, ok := result["value"].(string); ok {
		return value, nil
	}
	return "error", fmt.Errorf("Flipt response missing 'enabled' and 'value'")
}

func featureHandler(w http.ResponseWriter, r *http.Request) {
	fliptURL := os.Getenv("FLIPT_URL")
	if fliptURL == "" {
		fliptURL = defaultFliptURL
	}

	value, err := evaluateFeature(fliptURL)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<div id="feature-status" style="color:red;">Error: %v</div>`, err)
		return
	}

	// Log state changes with timestamp in the browser (using JS)
	fmt.Fprintf(w, renderFeatureStatusJS(value))
}

// Returns a script and log table, appends a new row only if state changed
func renderFeatureStatusJS(value string) string {
	color := "red"
	status := "DISABLED"
	if value == "true" {
		color = "green"
		status = "ENABLED"
	}
	return fmt.Sprintf(`
<div id="feature-status-log">
	<table id="status-log-table" style="width:100%%;border-collapse:collapse;">
		<thead><tr><th style='text-align:left;'>Time</th><th>Status</th></tr></thead>
		<tbody id="status-log-body"></tbody>
	</table>
</div>
<script>
(function() {
	var lastStatus = window.lastFeatureStatus;
	var newStatus = "%s";
	if (lastStatus !== newStatus) {
		var now = new Date().toLocaleTimeString();
		var row = '<tr><td>' + now + '</td><td style="color:%s;"><b>' + newStatus + '</b></td></tr>';
		document.getElementById('status-log-body').insertAdjacentHTML('beforeend', row);
		window.lastFeatureStatus = newStatus;
	}
})();
</script>
<span style="color: %s;">Feature is <b>%s</b></span>
`, status, color, color, status)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Flipt Demo</title>
		<script src="https://unpkg.com/htmx.org@1.9.10"></script>
	</head>
	<body>
		<h1>Feature Flag Status</h1>
		<div id="feature-status">Loading...</div>
		<script>
			document.addEventListener('DOMContentLoaded', function() {
				htmx.ajax('GET', '/feature-status', '#feature-status');
			});
		</script>
	</body>
	</html>
	`
	fmt.Fprintf(w, html)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/feature-status", featureHandler)
	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
