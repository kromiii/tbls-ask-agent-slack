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

## Deploy to k8s

Copy `config.yml.sample` to `config.yml` and edit it.

```sh
$ cp config.yml.sample config.yml

```

Create configmap and secret

```
kubectl create configmap tbls-schemas --from-file=config.yml
kubectl create secret generic tbls-ask-agent-slack --from-literal=slack-signing-secret=$SLACK_SIGNING_SECRET --from-literal=slack-oauth-token=$SLACK_OAUTH_TOKEN --from-literal=openai-api-key=$OPENAI_API_KEY --from-literal=github-token=$GITHUB_TOKEN
```

Build docker image locally

```
docker build -t tbls-ask-agent-slack:latest .
```

Apply manifests

```
kubectl apply -f manifests
```

## Slack-app settings

Please create an app using `manifests.yml` from https://api.slack.com/apps.
