package gemini

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
	"google.golang.org/api/option"
)

type ChatMessage struct {
	Role    string
	Content string
}

type Client struct {
	projectID   string
	location    string
	modelID     string
	client      *genai.Client
	chatSession *genai.ChatSession
}

func NewClient(projectID, location, modelID, keyFile string) (*Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, projectID, location, option.WithCredentialsFile(keyFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %v", err)
	}

	model := client.GenerativeModel(modelID)
	chatSession := model.StartChat()

	return &Client{
		projectID:   projectID,
		location:    location,
		modelID:     modelID,
		client:      client,
		chatSession: chatSession,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) ChatCompletion(messages []ChatMessage) (string, error) {
	if len(messages) == 0 {
		return "", errors.New("messages array is empty")
	}

	ctx := context.Background()

	// Convert ChatMessage array to genai.Content array and set chat history
	var contents []*genai.Content
	for _, msg := range messages {
		content := &genai.Content{
			Parts: []genai.Part{
				genai.Text(msg.Content),
			},
			Role: msg.Role,
		}
		contents = append(contents, content)
	}
	c.chatSession.History = contents

	// Send the last message
	lastMessage := messages[len(messages)-1]
	resp, err := c.chatSession.SendMessage(ctx, genai.Text(lastMessage.Content))
	if err != nil {
		return "", fmt.Errorf("failed to send message: %v", err)
	}

	return extractResponse(resp), nil
}

func extractResponse(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return ""
	}

	var response string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			response += string(textPart)
		}
	}
	return response
}
