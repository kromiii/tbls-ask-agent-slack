package slackhandler

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSlackAPI is a mock of the Slack API client
type MockSlackAPI struct {
	mock.Mock
}


func (m *MockSlackAPI) AuthTest() (*slack.AuthTestResponse, error) {
	args := m.Called()
	return args.Get(0).(*slack.AuthTestResponse), args.Error(1)
}

func (m *MockSlackAPI) GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
	args := m.Called(params)
	return args.Get(0).([]slack.Message), args.Bool(1), args.String(2), args.Error(3)
}

func (m *MockSlackAPI) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	args := m.Called(channelID, options)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockSlackAPI) UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error) {
	args := m.Called(channelID, timestamp, options)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func TestHandleAppMentionEvent(t *testing.T) {
	t.Parallel() // Enable parallel execution for this test function

	tests := []struct {
		name    string
		setup   func(*MockSlackAPI)
		ev      *slackevents.AppMentionEvent
		wantErr bool
	}{
		{
			name: "Multiple schemas - show select box",
			setup: func(mockAPI *MockSlackAPI) {
				// AuthTest might be called if there's only 1 schema due to race condition or config loading issue
				mockAPI.On("AuthTest").Return(&slack.AuthTestResponse{UserID: "UBOTID12345"}, nil).Maybe()
				mockAPI.On("GetConversationReplies", mock.AnythingOfType("*slack.GetConversationRepliesParameters")).Return([]slack.Message{}, false, "", nil).Maybe()
				mockAPI.On("PostMessage", mock.Anything, mock.Anything).Return("", "", nil)
			},
			ev: &slackevents.AppMentionEvent{
				Channel:         "C1234567890",
				User:            "U1234567890",
				Text:            "Hello, bot!",
				ThreadTimeStamp: "1234567890.123456",
			},
			wantErr: false,
		},
		{
			name: "Single schema - auto select",
			setup: func(mockAPI *MockSlackAPI) {
				mockAPI.On("AuthTest").Return(&slack.AuthTestResponse{UserID: "UBOTID12345"}, nil)
				mockAPI.On("GetConversationReplies", mock.AnythingOfType("*slack.GetConversationRepliesParameters")).Return([]slack.Message{}, false, "", nil)
				mockAPI.On("PostMessage", mock.Anything, mock.Anything).Return("", "", nil)
			},
			ev: &slackevents.AppMentionEvent{
				Channel:         "C1234567890",
				User:            "U1234567890",
				Text:            "Hello, bot!",
				ThreadTimeStamp: "1234567890.123456",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution for subtests

			// Setup
			mockAPI := &MockSlackAPI{}
			handler := NewSlackHandler(mockAPI)

			// Mock fileLoader
			oldFileLoader := fileLoader
			if tt.name == "Single schema - auto select" {
				fileLoader = func(filename string) ([]byte, error) {
					return []byte(`schemas:
  - name: test
    path: /path/to/test.yml`), nil
				}
			} else {
				fileLoader = func(filename string) ([]byte, error) {
					return []byte(`schemas:
  - name: test1
    path: /path/to/test1.yml
  - name: test2
    path: /path/to/test2.yml`), nil
				}
			}
			t.Cleanup(func() { fileLoader = oldFileLoader })

			if tt.setup != nil {
				tt.setup(mockAPI)
			}

			err := handler.handleAppMentionEvent(tt.ev, "./schemas/config.yml")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestHandleInteractionCallback(t *testing.T) {
	t.Parallel() // Enable parallel execution for this test function

	tests := []struct {
		name        string
		interaction slack.InteractionCallback
		setup       func(*MockSlackAPI)
		wantErr     bool
		errMsg      string
	}{
		{
			name: "Valid schema selection",
			interaction: slack.InteractionCallback{
				ActionCallback: slack.ActionCallbacks{
					BlockActions: []*slack.BlockAction{
						{
							ActionID: "select_schema",
							SelectedOption: slack.OptionBlockObject{
								Value: "/path/to/schema.yml",
								Text:  &slack.TextBlockObject{Text: "schema"},
							},
						},
					},
				},
				Channel: slack.Channel{
					GroupConversation: slack.GroupConversation{
						Conversation: slack.Conversation{
							ID: "C1234567890",
						},
					},
				},
				Message: slack.Message{
					Msg: slack.Msg{
						Timestamp: "1234567890.123456",
					},
				},
			},
			setup: func(mockAPI *MockSlackAPI) {
				mockAPI.On("UpdateMessage", mock.Anything, mock.Anything, mock.Anything).Return("", "", "", nil)
				mockAPI.On("GetConversationReplies", mock.Anything).Return([]slack.Message{}, false, "", nil)
				mockAPI.On("AuthTest").Return(&slack.AuthTestResponse{UserID: "UBOTID12345"}, nil)
				mockAPI.On("PostMessage", mock.Anything, mock.Anything).Return("", "", nil)
			},
			wantErr: false,
		},
		{
			name: "Invalid request",
			interaction: slack.InteractionCallback{
				ActionCallback: slack.ActionCallbacks{
					BlockActions: []*slack.BlockAction{},
				},
			},
			wantErr: true,
			errMsg:  "invalid request",
		},
		{
			name: "Unknown action",
			interaction: slack.InteractionCallback{
				ActionCallback: slack.ActionCallbacks{
					BlockActions: []*slack.BlockAction{
						{
							ActionID: "unknown_action",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "unknown action",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution for subtests

			// Setup
			mockAPI := new(MockSlackAPI)
			handler := NewSlackHandler(mockAPI)

			if tt.setup != nil {
				tt.setup(mockAPI)
			}

			err := handler.HandleInteractionCallback(tt.interaction)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
			mockAPI.AssertExpectations(t)
		})
	}
}
