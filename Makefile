.PHONY: create-configmap create-secret build-image apply-manifests clear all server embeddings

create-configmap:
	kubectl create configmap tbls-schemas --from-file=schemas/config.yml

create-secret:
	kubectl create secret generic tbls-ask-agent-slack --from-literal=slack-app-token=$$SLACK_APP_TOKEN --from-literal=slack-oauth-token=$$SLACK_OAUTH_TOKEN --from-literal=openai-api-key=$$OPENAI_API_KEY --from-literal=github-token=$$GITHUB_TOKEN

create-manual-embeddings-job:
	kubectl create job --from=cronjob/tbls-ask-bot-embeddings-job tbls-ask-bot-embeddings-manual-001

build-image:
	docker build -t tbls-ask-agent-slack:latest .

apply-manifests:
	kubectl apply -f manifests

clear:
	kubectl delete configmap tbls-schemas
	kubectl delete secret tbls-ask-agent-slack
	kubectl delete -f manifests

all: create-configmap create-secret apply-manifests

# For local development
server:
	go run main.go server

embeddings:
	go run main.go embeddings
