.PHONY: create-configmap create-secret build-image apply-manifests clear all server

create-configmap:
	kubectl create configmap tbls-schemas --from-file=schemas/config.yml

create-secret:
	kubectl create secret generic tbls-ask-agent-slack \
	--from-literal=slack-app-token=$$SLACK_APP_TOKEN \
	--from-literal=slack-oauth-token=$$SLACK_OAUTH_TOKEN \
	--from-literal=github-token=$$GITHUB_TOKEN \
	--from-literal=openai-api-key=$$OPENAI_API_KEY \
	--from-literal=google-application-credentials-json="$$GOOGLE_APPLICATION_CREDENTIALS_JSON"

build-image:
	docker build -t tbls-ask-agent-slack:latest .

apply-manifests:
	kubectl apply -f manifests/deployment.yml

clear:
	kubectl delete configmap tbls-schemas
	kubectl delete secret tbls-ask-agent-slack
	kubectl delete -f manifests/deployment.yml

all: create-configmap create-secret apply-manifests

# For local development
server:
	go run main.go server

