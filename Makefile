.PHONY: build-image helm-lint helm-template helm-deploy helm-uninstall server credits

build-image:
	docker build -t tbls-ask-agent-slack:latest .

helm-lint:
	helm lint charts/tbls-ask-agent-slack

helm-template:
	helm template tbls-ask-agent-slack ./charts/tbls-ask-agent-slack

helm-deploy:
	helm upgrade --install tbls-ask-agent-slack ./charts/tbls-ask-agent-slack \
		--set secret.slackAppToken="$$SLACK_APP_TOKEN" \
		--set secret.slackOAuthToken="$$SLACK_OAUTH_TOKEN" \
		--set secret.githubToken="$$GITHUB_TOKEN" \
		--set secret.openaiApiKey="$$OPENAI_API_KEY" \
		--set config.openaiBaseUrl="$$OPENAI_BASE_URL" \
		--set-file schemas.config=schemas/config.yml

helm-uninstall:
	helm uninstall tbls-ask-agent-slack

# For local development
server:
	go run main.go server

# Regenerate third-party dependency license notices
credits:
	go run github.com/Songmu/gocredits/cmd/gocredits@latest -skip-missing -w .
