package main

import (
	"net/http"
	"os"

	"github.com/kromiii/tbls-ask-agent-slack/handler"
	"github.com/kromiii/tbls-ask-agent-slack/slackhandler"
	"github.com/slack-go/slack"
)

func main() {
	// Slack Appの設定画面から取得する
	oauthToken := os.Getenv("SLACK_OAUTH_TOKEN")
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	api := slack.New(oauthToken)

	// SlackのEventおよびInteractionのハンドラ（再利用するため別定義）
	slackHandler := slackhandler.SlackHandler{
		Api: api,
	}

	// health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.Handle("/slack/events", &handler.EventHandler{
		SlackHandler:  &slackHandler,
		SigningSecret: signingSecret,
	})

	http.Handle("/slack/interaction", &handler.InteractivityHandler{
		SlackHandler: &slackHandler,
	})

	http.ListenAndServe(":8080", nil)
}
