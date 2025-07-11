package slackhandler

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/kromiii/tbls-ask-agent-slack/client"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"gopkg.in/yaml.v2"
)

type SlackHandler struct {
	Api           SlackAPI
	threadSchemas *ttlcache.Cache[string, *Schema] // threadTS -> selected schema with TTL
}

type Schema struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Config struct {
	Schemas []Schema `yaml:"schemas"`
}

type SlackAPI interface {
	GetConversationInfo(params *slack.GetConversationInfoInput) (*slack.Channel, error)
	AuthTest() (*slack.AuthTestResponse, error)
	GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error)
}

var fileLoader = os.ReadFile

// NewSlackHandler creates a new SlackHandler with initialized thread schema cache
func NewSlackHandler(api SlackAPI) *SlackHandler {
	cache := ttlcache.New[string, *Schema](
		ttlcache.WithTTL[string, *Schema](1 * time.Hour),
	)
	go cache.Start() // Start the cache's automatic cleanup goroutine

	return &SlackHandler{
		Api:           api,
		threadSchemas: cache,
	}
}

func (h *SlackHandler) HandleCallBackEvent(event slackevents.EventsAPIEvent, path string) error {
	innerEvent := event.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		return h.handleAppMentionEvent(ev, path)
	default:
		return errors.New("unknown event")
	}
}

func (h *SlackHandler) handleAppMentionEvent(ev *slackevents.AppMentionEvent, path string) error {
	// Check if schema is already selected for this thread
	threadTS := ev.ThreadTimeStamp
	if threadTS == "" {
		threadTS = ev.TimeStamp
	}

	if item := h.threadSchemas.Get(threadTS); item != nil {
		return h.handleMatchedSchema(ev, item.Value())
	}

	channelInfo, err := h.getChannelInfo(ev.Channel)
	if err != nil {
		return err
	}

	config, err := h.loadConfig(path)
	if err != nil {
		return err
	}

	matchedSchema := h.findMatchingSchema(channelInfo.Name, config.Schemas)

	if matchedSchema != nil {
		// Store the matched schema for this thread
		h.threadSchemas.Set(threadTS, matchedSchema, ttlcache.DefaultTTL)
		return h.handleMatchedSchema(ev, matchedSchema)
	} else {
		return h.handleUnmatchedSchema(ev, config.Schemas)
	}
}

func (h *SlackHandler) getChannelInfo(channelID string) (*slack.Channel, error) {
	channelInfo, err := h.Api.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get channel info: %w", err)
	}
	return channelInfo, nil
}

func (h *SlackHandler) loadConfig(path string) (*Config, error) {
	configBytes, err := fileLoader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func (h *SlackHandler) findMatchingSchema(channelName string, schemas []Schema) *Schema {
	for _, schema := range schemas {
		if strings.Contains(strings.ToLower(channelName), strings.ToLower(schema.Name)) {
			return &schema
		}
	}
	return nil
}

func (h *SlackHandler) handleMatchedSchema(ev *slackevents.AppMentionEvent, schema *Schema) error {
	botUserID, err := h.getBotUserID()
	if err != nil {
		return err
	}

	messages, threadTS, err := h.getMessages(ev)
	if err != nil {
		return err
	}

	model := os.Getenv("MODEL_NAME")
	if model == "" {
		model = "gpt-4o"
	}
	answer := client.Ask(messages, schema.Name, schema.Path, botUserID, model)

	return h.postMessage(ev.Channel, answer, threadTS)
}

func (h *SlackHandler) handleUnmatchedSchema(ev *slackevents.AppMentionEvent, schemas []Schema) error {
	// If there's only one schema, use it automatically. If there are no schemas, notify the user.
	if len(schemas) <= 1 {
		if len(schemas) == 0 {
			return h.postMessage(ev.Channel, "No schemas are configured.", ev.TimeStamp)
		}
		return h.handleMatchedSchema(ev, &schemas[0])
	}

	options := h.createSchemaOptions(schemas)

	_, _, err := h.Api.PostMessage(ev.Channel, slack.MsgOptionBlocks(
		slack.SectionBlock{
			Type: slack.MBTSection,
			Text: &slack.TextBlockObject{
				Type: slack.PlainTextType,
				Text: "Please select the target schema",
			},
			Accessory: &slack.Accessory{
				SelectElement: &slack.SelectBlockElement{
					ActionID: "select_schema",
					Type:     slack.OptTypeStatic,
					Placeholder: &slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "Select schema",
					},
					Options: options,
				},
			},
		},
	), slack.MsgOptionTS(ev.TimeStamp))

	return err
}

func (h *SlackHandler) getBotUserID() (string, error) {
	response, err := h.Api.AuthTest()
	if err != nil {
		return "", fmt.Errorf("failed to get bot user ID: %w", err)
	}
	return response.UserID, nil
}

func (h *SlackHandler) getMessages(ev *slackevents.AppMentionEvent) ([]slack.Message, string, error) {
	var messages []slack.Message
	var threadTS string
	var err error

	if ev.ThreadTimeStamp != "" {
		messages, _, _, err = h.Api.GetConversationReplies(
			&slack.GetConversationRepliesParameters{
				ChannelID: ev.Channel,
				Timestamp: ev.ThreadTimeStamp,
				Inclusive: true,
			},
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get conversation replies: %w", err)
		}
		threadTS = ev.ThreadTimeStamp
	} else {
		messages = []slack.Message{
			{
				Msg: slack.Msg{
					User:    ev.User,
					Text:    ev.Text,
					Channel: ev.Channel,
				},
			},
		}
		threadTS = ev.TimeStamp
	}

	return messages, threadTS, nil
}

func (h *SlackHandler) postMessage(channel, text, threadTS string) error {
	_, _, err := h.Api.PostMessage(
		channel,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(threadTS),
	)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}

func (h *SlackHandler) createSchemaOptions(schemas []Schema) []*slack.OptionBlockObject {
	var options []*slack.OptionBlockObject
	for _, schema := range schemas {
		options = append(options, &slack.OptionBlockObject{
			Text:  &slack.TextBlockObject{Type: slack.PlainTextType, Text: schema.Name},
			Value: schema.Path,
		})
	}
	return options
}

func (h *SlackHandler) HandleInteractionCallback(interaction slack.InteractionCallback) error {
	if len(interaction.ActionCallback.BlockActions) != 1 {
		return errors.New("invalid request")
	}

	action := interaction.ActionCallback.BlockActions[0]
	switch action.ActionID {
	case "select_schema":
		return h.handleSchemaSelection(interaction, action)
	default:
		return errors.New("unknown action")
	}
}

func (h *SlackHandler) handleSchemaSelection(interaction slack.InteractionCallback, action *slack.BlockAction) error {
	selectedPath := action.SelectedOption.Value
	selectedName := action.SelectedOption.Text.Text

	// Update the message to remove the form and show the selected schema
	updatedText := fmt.Sprintf("Selected schema: *%s*\n\nProcessing your query...", selectedName)
	_, _, _, err := h.Api.UpdateMessage(
		interaction.Channel.ID,
		interaction.Message.Timestamp,
		slack.MsgOptionText(updatedText, false),
	)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	// Store the selected schema for this thread
	threadTS := interaction.Message.ThreadTimestamp
	if threadTS == "" {
		threadTS = interaction.Message.Timestamp
	}

	selectedSchema := &Schema{
		Name: selectedName,
		Path: selectedPath,
	}

	h.threadSchemas.Set(threadTS, selectedSchema, ttlcache.DefaultTTL)

	// Get conversation replies to process the query
	messages, _, _, err := h.Api.GetConversationReplies(
		&slack.GetConversationRepliesParameters{
			ChannelID: interaction.Channel.ID,
			Timestamp: threadTS,
			Inclusive: true,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get conversation replies: %w", err)
	}

	botUserID, err := h.getBotUserID()
	if err != nil {
		return err
	}

	model := os.Getenv("MODEL_NAME")
	if model == "" {
		model = "gpt-4o"
	}

	answer := client.Ask(messages, selectedName, selectedPath, botUserID, model)

	return h.postMessage(interaction.Channel.ID, answer, threadTS)
}
