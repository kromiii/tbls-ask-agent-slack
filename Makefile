VALUES_FILE ?= my-values.yaml

.PHONY: build-image server credits helm-install helm-uninstall

build-image:
	docker build -t tbls-ask-agent-slack:latest .

# Deploy using values file
helm-install:
	@if [ ! -f $(VALUES_FILE) ]; then \
		echo "Error: '$(VALUES_FILE)' not found."; \
		echo "Please copy chart/values-openai.yaml.example (or values-gemini.yaml.example) to $(VALUES_FILE) and set your tokens."; \
		exit 1; \
	fi
	helm upgrade --install tbls-ask ./chart -f $(VALUES_FILE)

helm-uninstall:
	helm uninstall tbls-ask

# For local development
server:
	go run main.go server

# Regenerate third-party dependency license notices
credits:
	go run github.com/Songmu/gocredits/cmd/gocredits@latest -skip-missing -w .

