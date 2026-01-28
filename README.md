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

* OPENAI_API_KEY: API key for OpenAI

If you want to use Gemini models, you need to set `GEMINI_API_KEY`.

* GEMINI_API_KEY: API key for Gemini

### Optional
* GITHUB_TOKEN: Token for GitHub API (optional)
* CUSTOM_INSTRUCTION: Custom instruciton for LLM (optional)
* DEBUG_MODE: When set to "true", outputs prompt contents to logs (optional)

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
