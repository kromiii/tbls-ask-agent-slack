package slackhandler

import (
	"errors"
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

func (m *MockSlackAPI) GetConversationInfo(params *slack.GetConversationInfoInput) (*slack.Channel, error) {
	args := m.Called(params)
	return args.Get(0).(*slack.Channel), args.Error(1)
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

func TestHandleAppMentionEvent(t *testing.T) {
	t.Parallel() // Enable parallel execution for this test function

	tests := []struct {
		name    string
		setup   func(*MockSlackAPI)
		ev      *slackevents.AppMentionEvent
		wantErr bool
	}{
		{
			name: "Matched schema",
			setup: func(mockAPI *MockSlackAPI) {
				mockAPI.On("GetConversationInfo", mock.Anything).Return(&slack.Channel{
					GroupConversation: slack.GroupConversation{
						Name:         "test-channel",
						Conversation: slack.Conversation{},
					},
				}, nil)
				mockAPI.On("AuthTest").Return(&slack.AuthTestResponse{UserID: "UBOTID12345"}, nil)
				mockAPI.On("GetConversationReplies", mock.Anything).Return([]slack.Message{}, false, "", nil)
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
			name: "Unmatched schema",
			setup: func(mockAPI *MockSlackAPI) {
				mockAPI.On("GetConversationInfo", mock.Anything).Return(&slack.Channel{
					GroupConversation: slack.GroupConversation{
						Name:         "other-channel",
						Conversation: slack.Conversation{},
					},
				}, nil)
				mockAPI.On("AuthTest").Return(&slack.AuthTestResponse{UserID: "UBOTID12345"}, nil)
				mockAPI.On("GetConversationReplies", mock.Anything).Return([]slack.Message{}, false, "", nil)
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
			name: "Error getting channel info",
			setup: func(mockAPI *MockSlackAPI) {
				mockAPI.On("GetConversationInfo", mock.Anything).Return((*slack.Channel)(nil), errors.New("API error"))
			},
			ev: &slackevents.AppMentionEvent{
				Channel:         "C1234567890",
				User:            "U1234567890",
				Text:            "Hello, bot!",
				ThreadTimeStamp: "1234567890.123456",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution for subtests

			// Setup
			mockAPI := &MockSlackAPI{}
			handler := &SlackHandler{Api: mockAPI}

			// Mock fileLoader
			oldFileLoader := fileLoader
			fileLoader = func(filename string) ([]byte, error) {
				return []byte(`schemas:
  - name: test
    path: /path/to/test.yml`), nil
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
			handler := &SlackHandler{Api: mockAPI}

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
