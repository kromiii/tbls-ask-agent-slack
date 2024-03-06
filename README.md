# tbls-ask-agent-slack

@k1Low の [tbls-ask](https://github.com/k1LoW/tbls-ask) を slack-bot にしたものです

メンションして話しかけることでslackからtbls-askを呼び出せます

## デモ

https://github.com/kromiii/tbls-ask-agent-slack/assets/15026387/b6ff5027-5af3-4e21-b95e-23584506bcbe

## 環境変数

* OPENAI_API_KEY: OpenAIのAPIキー
* SLACK_APP_TOKEN: SlackのAppトークン
* SLACK_OAUTH_TOKEN: SlackのOAuthトークン

## k8s へのデプロイ

```
$ kubectl create configmap tbls-schemas --from-file=config.yml
$ kubectl create secret generic tbls-ask-agent-slack --from-literal=slack-signing-secret=$SLACK_SIGNING_SECRET --from-literal=slack-oauth-token=$SLACK_OAUTH_TOKEN --from-literal=openai-api-key=$OPENAI_API_KEY
$ kubectl apply -f manifests
```
