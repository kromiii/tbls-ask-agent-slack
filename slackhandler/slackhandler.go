package slackhandler

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/kromiii/tbls-ask-agent-slack/client"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"gopkg.in/yaml.v2"
)

type SlackHandler struct {
	Api           SlackAPI
	threadSchemas *ttlcache.Cache[string, []*client.SchemaInfo] // threadTS -> selected schemas with TTL
	configPath    string                                        // config file path for loading schemas
}

type Schema struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Config struct {
	Schemas []Schema `yaml:"schemas"`
}

type SlackAPI interface {
	AuthTest() (*slack.AuthTestResponse, error)
	GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error)
}

var fileLoader = os.ReadFile

// NewSlackHandler creates a new SlackHandler with initialized thread schema cache
func NewSlackHandler(api SlackAPI) *SlackHandler {
	cache := ttlcache.New[string, []*client.SchemaInfo](
		ttlcache.WithTTL[string, []*client.SchemaInfo](1 * time.Hour),
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

	h.configPath = path

	if item := h.threadSchemas.Get(threadTS); item != nil {
		return h.processQueryWithKnownSchemas(ev, item.Value())
	}

	config, err := h.loadConfig(path)
	if err != nil {
		return err
	}

	return h.promptSchemaSelection(ev, config.Schemas)
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

func (h *SlackHandler) processQueryWithKnownSchemas(ev *slackevents.AppMentionEvent, schemas []*client.SchemaInfo) error {
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
	answer := client.AskWithSchemas(messages, schemas, botUserID, model)

	return h.postMessage(ev.Channel, answer, threadTS)
}

func (h *SlackHandler) promptSchemaSelection(ev *slackevents.AppMentionEvent, schemas []Schema) error {
	// If there's only one schema, use it automatically. If there are no schemas, notify the user.
	if len(schemas) <= 1 {
		if len(schemas) == 0 {
			return h.postMessage(ev.Channel, "No schemas are configured.", ev.TimeStamp)
		}
		return h.processQueryWithKnownSchemas(ev, []*client.SchemaInfo{{Name: schemas[0].Name, Path: schemas[0].Path}})
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

const allSchemaValue = "__all__"

func (h *SlackHandler) createSchemaOptions(schemas []Schema) []*slack.OptionBlockObject {
	var options []*slack.OptionBlockObject

	// Add "all" option first when there are multiple schemas
	if len(schemas) > 1 {
		options = append(options, &slack.OptionBlockObject{
			Text:  &slack.TextBlockObject{Type: slack.PlainTextType, Text: "all"},
			Value: allSchemaValue,
		})
	}

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

	var selectedSchemas []*client.SchemaInfo

	// Handle "all" selection
	if selectedPath == allSchemaValue {
		config, err := h.loadConfig(h.configPath)
		if err != nil {
			return fmt.Errorf("failed to load config for all schemas: %w", err)
		}
		for _, s := range config.Schemas {
			selectedSchemas = append(selectedSchemas, &client.SchemaInfo{Name: s.Name, Path: s.Path})
		}
	} else {
		selectedSchemas = []*client.SchemaInfo{{
			Name: selectedName,
			Path: selectedPath,
		}}
	}

	h.threadSchemas.Set(threadTS, selectedSchemas, ttlcache.DefaultTTL)

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

	answer := client.AskWithSchemas(messages, selectedSchemas, botUserID, model)

	return h.postMessage(interaction.Channel.ID, answer, threadTS)
}
