package chat

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type ChatReq struct {
	Model       string    `json:"model,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	MaxTokens   int64     `json:"max_tokens"`
}

type Message struct {
	Role    string
	Content string
}

type ChatResp struct {
	Id      string
	Object  string
	Created int64
	Model   string
	Choices []Choice
	Usage   Usage
}

type Choice struct {
	Index        int64
	Message      Message
	FinishReason string `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

func (chat *Chat) chat(cmd *ChatCommand) (res *ChatResp, err error) {
	chatReq := &ChatReq{
		Messages: []Message{
			{Role: "system", Content: "You are an AI assistant that helps people find information."},
			{Role: "user", Content: cmd.Content},
		},
		Temperature: 0.7,
		TopP:        0.95,
		MaxTokens:   800,
	}
	data, err := json.Marshal(chatReq)
	if err != nil {
		return
	}
	req, _ := http.NewRequest(http.MethodPost, "https://cybever-openai.openai.azure.com/openai/deployments/hty/chat/completions?api-version=2024-02-15-preview", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", chat.conf.APIKey)
	resp, err := chat.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res = &ChatResp{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	return
}
