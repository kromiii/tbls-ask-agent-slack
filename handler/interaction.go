package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kromiii/tbls-ask-agent-slack/slackhandler"
	"github.com/slack-go/slack"
)

type InteractivityHandler struct {
	SlackHandler *slackhandler.SlackHandler
	SlackClient  *slack.Client
}

func (h *InteractivityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var interaction slack.InteractionCallback
	err := json.Unmarshal([]byte(r.FormValue("payload")), &interaction)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.SlackHandler.HandleInteractionCallback(interaction)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
