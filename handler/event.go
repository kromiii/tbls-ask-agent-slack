package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/kromiii/tbls-ask-agent-slack/slackhandler"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type EventHandler struct {
	SlackHandler  *slackhandler.SlackHandler
	SigningSecret string
}

func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// リクエストの検証
	sv, err := slack.NewSecretsVerifier(r.Header, h.SigningSecret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := sv.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := sv.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// eventをパース
	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// URLVerification eventをhandle（EventAPI有効化時に叩かれる）
	if event.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	}

	// その他Eventのハンドリング（以下、slackhandler.SlackHandlerで定義）
	if event.Type == slackevents.CallbackEvent {
		err := h.SlackHandler.HandleCallBackEvent(event)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
