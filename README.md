# tbls-ask-agent-slack

This is a slack-bot version of [tbls-ask](https://github.com/k1LoW/tbls-ask).

You can call tbls-ask from slack by mentioning the bot.

<img width="869" alt="image" src="https://github.com/user-attachments/assets/1a0657d8-16a4-4f8f-b040-80a3093e36c2">

# Usage

To use the tbls-ask-agent-slack bot, mention it in a Slack channel where it's been added. The bot can answer questions about your database schema using natural language.

### Example Query

You can ask questions like:

"@tbls-ask name of the users who has the most stars"

The bot will respond with an SQL query that answers your question. For the example above, it would provide a query to find the user with the most stars on their comments.

## How It Works

1. The bot receives your question through a Slack mention.
2. It interprets your natural language query.
3. It generates an appropriate SQL query to answer your question.
4. The bot returns the SQL query in the Slack thread.

### Multiple Schemas Support

If you have configured multiple schemas in `schemas/config.yml`, the bot will ask you to select a schema (or "all") when you mention it.

1. Mention the bot with your question.
2. The bot will respond with a dropdown menu to select the target schema.
3. Select the schema you want to query against.
4. The bot will generate the SQL query based on the selected schema.

## Limitations

- The bot provides SQL queries but does not execute them directly on your database.
- You need to run the provided SQL query on your own MySQL database to get the actual results.
- The database document must be prepared using [`tbls`](https://github.com/k1LoW/tbls).

# How to set up

## Environment Variables

### Required

* SLACK_APP_TOKEN: App token for Slack
* SLACK_OAUTH_TOKEN: OAuth token for Slack
* MODEL_NAME: Model name for LLM (default: gpt-4o)

### API keys
By default, we use OpenAI models. You need to set `OPENAI_API_KEY`.

* OPENAI_API_KEY: API key for OpenAI (or OpenAI-compatible API)

### Optional
* OPENAI_BASE_URL: Base URL for OpenAI-compatible endpoints (e.g. `https://generativelanguage.googleapis.com/v1beta/openai/` for Gemini, or Ollama/OpenRouter) (optional)
* GITHUB_TOKEN: Token for GitHub API (optional)
* CUSTOM_INSTRUCTION: Custom instruction for LLM (optional)
* DEBUG_MODE: When set to "true", outputs prompt contents to logs (optional)

### Setup .env (For local development)

Copy `.env.sample` to `.env` and edit it with your credentials:

```sh
$ cp .env.sample .env
```

## Slack-app settings

Please create an app using `manifest.yml` and install it to your workspace.

## Prepare schema

Copy `schemas/config.yml.sample` to `schemas/config.yml` and edit it. You can define multiple schemas in this file.

```sh
$ cp schemas/config.yml.sample schemas/config.yml
```

## Run Server

```sh
make server
```

This app uses socket mode for slack, so you don't need to expose the server to the internet. That means you don't need to set `SLACK_SIGNING_SECRET`.

## Deploy to k8s

You can deploy `tbls-ask-agent-slack` to Kubernetes using the Helm chart located in `charts/tbls-ask-agent-slack`.

### Quick start with Makefile

1. Build docker image locally (if needed):

```sh
make build-image
```

2. Deploy using Helm (uses environment variables and `schemas/config.yml`):

```sh
make helm-deploy
```

To uninstall:

```sh
make helm-uninstall
```

### Deploy directly with Helm

```sh
helm upgrade --install tbls-ask-agent-slack ./charts/tbls-ask-agent-slack \
  --set secret.slackAppToken="$SLACK_APP_TOKEN" \
  --set secret.slackOAuthToken="$SLACK_OAUTH_TOKEN" \
  --set secret.openaiApiKey="$OPENAI_API_KEY" \
  --set secret.githubToken="$GITHUB_TOKEN" \
  --set config.openaiBaseUrl="$OPENAI_BASE_URL" \
  --set-file schemas.config=schemas/config.yml
```

Or customize settings with a custom `values.yaml`:

```sh
helm upgrade --install tbls-ask-agent-slack ./charts/tbls-ask-agent-slack -f my-values.yaml
```

See [values.yaml](charts/tbls-ask-agent-slack/values.yaml) for all configurable parameters, including using existing Secrets / ConfigMaps and resource limits.

## License

This project is licensed under the [MIT License](LICENSE).

Third-party dependency licenses are listed in [`CREDITS`](CREDITS), generated with [gocredits](https://github.com/Songmu/gocredits) (run `make credits` to regenerate).
