SHELL := /bin/bash

.PHONY: help setup run build test docker-build up down clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

setup: ## Copy .env and tidy Go modules
	@test -f .env || cp .env.example .env
	go mod tidy

run: ## Run the agent REPL
	go run .

build: ## Build a local binary at bin/agent
	go build -o bin/agent .

test: ## Run unit tests
	go test ./...

docker-build: ## Build the Docker image
	docker build -t coding-agents-from-scratch-go:latest .

up: ## Start the container via docker compose
	docker compose up

down: ## Stop the container
	docker compose down

clean: ## Remove build artifacts
	rm -rf bin /tmp/lwp-coding-agent
