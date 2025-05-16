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
	flagKey         = "feature_toggle" // Boolean flag for enabling/disabling the feature
	colorFlagKey    = "color_box"      // Variant flag for color
	entityID        = "user123"
)

// Evaluate the feature flag, supporting Boolean flags only
func evaluateFeature(fliptURL string) (string, error) {
	url := fmt.Sprintf("%s/evaluate/v1/boolean", fliptURL)
	payload := map[string]interface{}{
		"namespaceKey": "default",
		"flagKey":      flagKey,
		"entityId":     entityID,
		"context":      map[string]interface{}{},
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

	var result struct {
		Enabled bool   `json:"enabled"`
		Reason  string `json:"reason"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "error", fmt.Errorf("failed to decode Flipt response: %w", err)
	}

	if result.Enabled {
		return "true", nil
	}
	return "false", nil
}

// Evaluate the color variant flag
func evaluateColor(fliptURL string) (string, error) {
	url := fmt.Sprintf("%s/evaluate/v1/variant", fliptURL)
	payload := map[string]interface{}{
		"namespaceKey": "default",
		"flagKey":      colorFlagKey,
		"entityId":     entityID,
		"context":      map[string]interface{}{},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "#888888", fmt.Errorf("failed to marshal JSON payload: %w", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "#888888", fmt.Errorf("failed to make request to Flipt: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "#888888", fmt.Errorf("Flipt returned non-OK status: %d", resp.StatusCode)
	}
	var result struct {
		VariantKey string `json:"variantKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "#888888", fmt.Errorf("failed to decode Flipt response: %w", err)
	}
	if result.VariantKey != "" {
		return result.VariantKey, nil
	}
	return "#888888", nil // fallback color
}

// Change featureHandler to return JSON for status and color
func featureHandler(w http.ResponseWriter, r *http.Request) {
	// Get Flipt URL from env or use default
	fliptURL := os.Getenv("FLIPT_URL")
	if fliptURL == "" {
		fliptURL = defaultFliptURL
	}

	value, err := evaluateFeature(fliptURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ERROR",
			"color":   "red",
			"warning": err.Error(),
		})
		return
	}
	status := "DISABLED"
	color := "red"
	warning := ""
	if value == "true" {
		status = "ENABLED"
		color = "green"
	} else if value != "false" {
		// Edge case: unexpected value
		warning = "Unexpected flag value: '" + value + "' (expected 'true' or 'false'). Feature is treated as DISABLED."
	}
	color, _ = evaluateColor(fliptURL) // ignore error, fallback handled
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   status,
		"color":    color,
		"warning":  warning,
		"boxColor": color,
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
			#color-box { width: 40px; height: 40px; border-radius: 8px; display: inline-block; margin-left: 2em; border: 2px solid #444; vertical-align: middle; }
		</style>
	</head>
	<body>
		<div class="header-row">
			<span class="header-title">Feature Flag Status</span>
			<span id="feature-status"></span>
			<span id="color-box"></span>
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
			let lastBoxColor = null;
			function pollStatus() {
				fetch('/feature-status').then(r => r.json()).then(data => {
					const status = data.status;
					const color = data.color;
					const warning = data.warning || "";
					const boxColor = data.boxColor || "#888888";
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
					const warnDiv = document.getElementById('feature-warning');
					if (warning) {
						if (!warnDiv) {
							const div = document.createElement('div');
							div.id = 'feature-warning';
							div.style.color = 'orange';
							div.style.marginTop = '1em';
							div.textContent = warning;
							document.body.insertBefore(div, document.getElementById('feature-status-log'));
						} else {
							warnDiv.textContent = warning;
						}
					} else if (warnDiv) {
						warnDiv.remove();
					}
					if (lastBoxColor !== boxColor) {
						document.getElementById('color-box').style.background = boxColor;
						lastBoxColor = boxColor;
					}
				});
			}
			pollStatus();
			setInterval(pollStatus, 1000);
		})();
		</script>
	</body>
	</html>
	`
	fmt.Fprint(w, html)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/feature-status", featureHandler)
	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
