package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GptResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

func ReqiestGPT(messqge string) string {
	url := "https://api.openai.com/v1/chat/completions"
	method := "POST"

	text := fmt.Sprintf(`{
		"model": "gpt-3.5-turbo",
		"messages": [{"role": "user", "content": "%s"}],
		"temperature": 0.7
	}`, messqge)
	payload := strings.NewReader(text)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Authorization", "Bearer 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")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var gptRes GptResponse
	fmt.Println("gpt response ", string(body))
	err = json.Unmarshal(body, &gptRes)
	if err != nil {
		fmt.Println("gpt ===== ", err)
		return ""
	}

	if len(gptRes.Choices) > 0 {
		return gptRes.Choices[0].Message.Content
	}

	return ""
}
