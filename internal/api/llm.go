package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Response struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

var ApiKey string

func AskWithSmallModel(msgList []Message) (string, error) {
	return askWithChatGPT("gpt-4o-mini", msgList)
}

func AskWithLargeModel(msgList []Message) (string, error) {
	if ApiKey == "" || len(ApiKey) < 20 {
		return "", fmt.Errorf("API key is not set")
	}
	return askWithChatGPT("gpt-4o", msgList)
}

func askWithChatGPT(model string, msgList []Message) (string, error) {
	body, err := json.Marshal(Request{
		Model:    model,
		Messages: msgList,
		Stream:   true,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ApiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("API Error (Status %d): %s", res.StatusCode, string(bodyBytes))
	}

	var result strings.Builder
	reader := bufio.NewReader(res.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		lineStr := string(line)
		if strings.TrimSpace(lineStr) == "" {
			continue
		}

		if !strings.HasPrefix(lineStr, "data: ") {
			continue
		}
		jsonData := strings.TrimPrefix(strings.TrimSpace(lineStr), "data: ")

		if jsonData == "[DONE]" {
			continue
		}

		var stream Response
		if err := json.Unmarshal([]byte(jsonData), &stream); err != nil {
			continue
		}

		if len(stream.Choices) > 0 {
			content := stream.Choices[0].Delta.Content
			if content != "" {
				result.WriteString(content)
			}
		}
	}

	return result.String(), nil
}
