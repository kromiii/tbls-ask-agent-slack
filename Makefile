.PHONY: build-image server credits helm-install-openai helm-install-gemini helm-uninstall

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

