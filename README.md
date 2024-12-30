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

* OPENAI_API_KEY: API key for OpenAI

If you want to use Gemini models, you need to set either `GEMINI_API_KEY` or `GOOGLE_APPLICATION_CREDENTIALS_JSON`.

* GEMINI_API_KEY: API key for Gemini
* GOOGLE_APPLICATION_CREDENTIALS_JSON: JSON key for Google Cloud

### Optional
* GITHUB_TOKEN: Token for GitHub API (optional)
* CUSTOM_INSTRUCTION: Custom instruciton for LLM (optional)
* GOOGLE_CLOUD_REGION: Region for Google Cloud (optional, default: us-central1)

## Slack-app settings

Please create an app using `manifests.yml` and install it to your workspace.

## Prepare schema

Copy `schemas/config.yml.sample` to `schemas/config.yml` and edit it.

```sh
$ cp schemas/config.yml.sample schemas/config.yml
```

## Run Server

```sh
make server
```

It is using socket mode for slack. You don't need to expose the server to the internet.

## Deploy to k8s

Build docker image locally

```
make build-image
```

Choose which API to use

```
cp manifests/deployment-openai.yml manifests/deployment.yml
```

Apply manifests

```
make all
```
