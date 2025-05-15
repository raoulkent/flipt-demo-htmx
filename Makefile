# Makefile for Flipt + Go HTMX Demo

SHELL := /bin/bash

# Use Minikube's Docker daemon for image builds
minikube-docker-env:
	eval $$(minikube docker-env)

.PHONY: all build deploy redeploy clean port-forward

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
