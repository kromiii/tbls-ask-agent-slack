# tbls-ask-agent-slack

This is a slack-bot version of [tbls-ask](https://github.com/k1LoW/tbls-ask).

You can call tbls-ask from slack by mentioning the bot.

<img width="869" alt="image" src="https://github.com/user-attachments/assets/1a0657d8-16a4-4f8f-b040-80a3093e36c2">


## Environment Variables

* OPENAI_API_KEY: API key for OpenAI
* SLACK_APP_TOKEN: App token for Slack
* SLACK_OAUTH_TOKEN: OAuth token for Slack
* GITHUB_TOKEN: Token for GitHub API (optional)

## Slack-app settings

Please create an app using `manifests.yml` and install it to your workspace.

## Prepare schema

Copy `schemas/config.yml.sample` to `schemas/config.yml` and edit it.

```sh
$ cp schemas/config.yml.sample schemas/config.yml
```

## Run locally

```sh
go run main.go
```

It is using socket mode for slack. You don't need to expose the server to the internet.

## Deploy to k8s

Build docker image locally

```
make build-image
```

Apply manifests

```
make all
```
