.PHONY: create-configmap create-secret build-image apply-manifests clear all server credits helm-install-openai helm-install-gemini helm-uninstall

create-configmap:
	kubectl create configmap tbls-schemas --from-file=schemas/config.yml

create-secret:
	kubectl create secret generic tbls-ask-agent-slack \
	--from-literal=slack-app-token=$$SLACK_APP_TOKEN \
	--from-literal=slack-oauth-token=$$SLACK_OAUTH_TOKEN \
	--from-literal=github-token=$$GITHUB_TOKEN \
	--from-literal=openai-api-key=$$OPENAI_API_KEY \
	--from-literal=gemini-api-key=$$GEMINI_API_KEY

build-image:
	docker build -t tbls-ask-agent-slack:latest .

apply-manifests:
	kubectl apply -f manifests/deployment.yml

clear:
	kubectl delete configmap tbls-schemas
	kubectl delete secret tbls-ask-agent-slack
	kubectl delete -f manifests/deployment.yml

all: create-configmap create-secret apply-manifests

# Helm deployment targets
helm-install-openai:
	helm upgrade --install tbls-ask ./chart \
		--set provider=openai \
		--set modelName=gpt-4o-mini \
		--set secret.slackAppToken=$$SLACK_APP_TOKEN \
		--set secret.slackOAuthToken=$$SLACK_OAUTH_TOKEN \
		--set secret.openaiApiKey=$$OPENAI_API_KEY \
		--set secret.githubToken=$$GITHUB_TOKEN

helm-install-gemini:
	helm upgrade --install tbls-ask ./chart \
		--set provider=gemini \
		--set modelName=gemini-1.5-pro \
		--set secret.slackAppToken=$$SLACK_APP_TOKEN \
		--set secret.slackOAuthToken=$$SLACK_OAUTH_TOKEN \
		--set secret.geminiApiKey=$$GEMINI_API_KEY \
		--set secret.githubToken=$$GITHUB_TOKEN

helm-uninstall:
	helm uninstall tbls-ask

# For local development
server:
	go run main.go server

# Regenerate third-party dependency license notices
credits:
	go run github.com/Songmu/gocredits/cmd/gocredits@latest -skip-missing -w .

