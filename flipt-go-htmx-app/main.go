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

func evaluateFeature(fliptURL string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/evaluate", fliptURL)
	payload := map[string]interface{}{
		"flagKey":  flagKey,
		"entityId": entityID,
		"context":  map[string]interface{}{},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	// Create a bytes.Buffer from the byte slice
	body := bytes.NewBuffer(payloadBytes)

	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return false, fmt.Errorf("failed to make request to Flipt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Flipt returned non-OK status: %d", resp.StatusCode)
	}

	var evaluation struct {
		Enabled bool `json:"enabled"`
	}
	err = json.NewDecoder(resp.Body).Decode(&evaluation)
	if err != nil {
		return false, fmt.Errorf("failed to decode Flipt response: %w", err)
	}

	return evaluation.Enabled, nil
}

func featureHandler(w http.ResponseWriter, r *http.Request) {
	fliptURL := os.Getenv("FLIPT_URL")
	if (fliptURL == "") {
		fliptURL = defaultFliptURL
	}

	enabled, err := evaluateFeature(fliptURL)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		// Show error in the UI for easier debugging
		fmt.Fprintf(w, `<div id="feature-status" style="color:red;">Error: %v</div>`, err)
		return
	}

	fmt.Fprintf(w, `<div id="feature-status">%s</div>`, renderFeatureStatus(enabled))
}

func renderFeatureStatus(enabled bool) string {
	if enabled {
		return `<span style="color: green;">Feature is <b hx-swap="outerHTML" hx-get="/feature-status" hx-trigger="every 2s">ENABLED</b></span>`
	}
	return `<span style="color: red;">Feature is <b hx-swap="outerHTML" hx-get="/feature-status" hx-trigger="every 2s">DISABLED</b></span>`
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
