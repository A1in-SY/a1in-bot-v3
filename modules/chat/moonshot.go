package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type MoonshotChatReq struct {
	Model       string
	Messages    []Message
	Temperature float64
}

type Message struct {
	Role    string
	Content string
}

type MoonshotChatResp struct {
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

func (chat *Chat) chat(cmd *ChatCommand) (res *MoonshotChatResp, err error) {
	chatReq := &MoonshotChatReq{
		Model:       "moonshot-v1-8k",
		Messages:    []Message{{Role: "user", Content: cmd.Content}},
		Temperature: 0.3,
	}
	data, err := json.Marshal(chatReq)
	if err != nil {
		return
	}
	req, _ := http.NewRequest(http.MethodPost, "https://api.moonshot.cn/v1/chat/completions", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %v", chat.conf.APIKey))
	resp, err := chat.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res = &MoonshotChatResp{}
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	return
}
