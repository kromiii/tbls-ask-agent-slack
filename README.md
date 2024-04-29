# tbls-ask-agent-slack

This is a slack-bot version of [tbls-ask](https://github.com/k1LoW/tbls-ask).

You can call tbls-ask from slack by mentioning the bot.

## Demo

https://github.com/kromiii/tbls-ask-agent-slack/assets/15026387/b6ff5027-5af3-4e21-b95e-23584506bcbe

## Environment Variables

* OPENAI_API_KEY: API key for OpenAI
* SLACK_APP_TOKEN: App token for Slack
* SLACK_OAUTH_TOKEN: OAuth token for Slack
* GITHUB_TOKEN: Token for GitHub API (optional)

## Slack-app settings

Please create an app using `manifests.yml` from https://api.slack.com/apps.

## Deploy to k8s

Copy `config.yml.sample` to `config.yml` and edit it.

```sh
$ cp config.yml.sample config.yml

```

Build docker image locally

```
make build-image
```

Apply manifests

```
make all
```
