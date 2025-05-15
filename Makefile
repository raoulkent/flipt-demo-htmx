# Makefile for Flipt + Go HTMX Demo

SHELL := /bin/bash

# Use Minikube's Docker daemon for image builds
minikube-docker-env:
	eval $$(minikube docker-env)

.PHONY: all build deploy redeploy clean port-forward port-forward-flipt urls logs help

all: build deploy

build:
	@echo "[+] Building Go app Docker image inside Minikube Docker environment..."
	eval $$(minikube docker-env) && cd flipt-go-htmx-app && docker build -t flipt-go-htmx-app:latest .

# Deploy both the Go app and Flipt to Kubernetes
# Assumes k8s/deployment.yaml and k8s/service.yaml are correct
# Will also update services if needed

deploy:
	@echo "[+] Deploying to Kubernetes..."
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml

redeploy: build deploy
	kubectl rollout restart deployment/flipt-go-htmx-app

clean:
	@echo "[+] Deleting all demo resources from Kubernetes..."
	kubectl delete -f k8s/service.yaml --ignore-not-found
	kubectl delete -f k8s/deployment.yaml --ignore-not-found

# Port-forward Go app to localhost:8081 (for fixed port access)
port-forward:
	kubectl port-forward deployment/flipt-go-htmx-app 8081:8080

# Port-forward Flipt dashboard to localhost:8080
port-forward-flipt:
	kubectl port-forward deployment/flipt 8080:8080

# Print URLs for both services (using minikube service)
urls:
	@echo "Flipt dashboard URL:"
	minikube service flipt --url
	@echo "Go HTMX app URL:"
	minikube service flipt-go-htmx-app-service --url

# View Go app logs
logs:
	kubectl logs deployment/flipt-go-htmx-app -f

# Show all Makefile targets
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-24s %s\n", $$1, $$2}'
	@echo "\nCommon targets: all, build, deploy, redeploy, clean, port-forward, port-forward-flipt, urls, logs, help"
