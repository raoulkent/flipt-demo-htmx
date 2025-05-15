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

// Change featureHandler to return JSON for status and color
func featureHandler(w http.ResponseWriter, r *http.Request) {
	fliptURL := os.Getenv("FLIPT_URL")
	if fliptURL == "" {
		fliptURL = defaultFliptURL
	}

	value, err := evaluateFeature(fliptURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ERROR",
			"color": "red",
		})
		return
	}
	status := "DISABLED"
	color := "red"
	if value == "true" {
		status = "ENABLED"
		color = "green"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": status,
		"color": color,
	})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Flipt Demo</title>
		<script src="https://unpkg.com/htmx.org@1.9.10"></script>
		<style>
			body { background: #18191a; color: #f5f6fa; font-family: sans-serif; }
			table { background: #232526; }
			#status-log-table { max-height: 300px; overflow-y: auto; display: block; }
			.header-row { display: flex; align-items: center; margin-bottom: 1.5em; }
			.header-title { font-size: 2.5em; font-weight: bold; margin-right: 1.5em; }
			#feature-status { font-size: 1.5em; margin-left: 1em; }
		</style>
	</head>
	<body>
		<div class="header-row">
			<span class="header-title">Feature Flag Status</span>
			<span id="feature-status"></span>
		</div>
		<div id="feature-status-log">
			<table id="status-log-table" style="width:100%;border-collapse:collapse;">
				<thead><tr><th style='text-align:left;'>Time</th><th>Status</th></tr></thead>
				<tbody id="status-log-body"></tbody>
			</table>
		</div>
		<script>
		(function() {
			let lastStatus = null;
			function pollStatus() {
				fetch('/feature-status').then(r => r.json()).then(data => {
					const status = data.status;
					const color = data.color;
					const now = new Date().toLocaleTimeString();
					if (lastStatus !== status) {
						var row = document.createElement('tr');
						row.appendChild(Object.assign(document.createElement('td'), {textContent: now}));
						var statusCell = document.createElement('td');
						statusCell.style.color = color;
						statusCell.innerHTML = '<b>' + status + '</b>';
						row.appendChild(statusCell);
						document.getElementById('status-log-body').appendChild(row);
						row.scrollIntoView({behavior: 'smooth', block: 'end'});
						lastStatus = status;
					}
					document.getElementById('feature-status').innerHTML = '<span style="color:' + color + ';">Feature is <b>' + status + '</b></span>';
				});
			}
			pollStatus();
			setInterval(pollStatus, 1000);
		})();
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
