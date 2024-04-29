.PHONY: create-configmap create-secret build-image apply-manifests clear all

create-configmap:
	kubectl create configmap tbls-schemas --from-file=config.yml

create-secret:
	kubectl create secret generic tbls-ask-agent-slack --from-literal=slack-app-token=$$SLACK_APP_TOKEN --from-literal=slack-oauth-token=$$SLACK_OAUTH_TOKEN --from-literal=openai-api-key=$$OPENAI_API_KEY --from-literal=github-token=$$GITHUB_TOKEN

build-image:
	docker build -t tbls-ask-agent-slack:latest .

apply-manifests:
	kubectl apply -f manifests

clear:
	kubectl delete configmap tbls-schemas
	kubectl delete secret tbls-ask-agent-slack
	kubectl delete -f manifests

all: create-configmap create-secret apply-manifests
