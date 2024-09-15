package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type openai struct {
	apiKey  string
	baseURL string
}

func NewOpenAI(apiKey string) *openai {
	return &openai{apiKey, "https://api.openai.com/v1"}
}

func (oai *openai) Chat(ctx context.Context, model string, temperature int32, messages []Message, tools []Tool) (<-chan Message, <-chan string, error) {
	rb := openAIReqBody{
		Messages:  messages,
		MaxTokens: 4096,
		Model:     model,
		Stream:    false,
	}

	rbBytes, err := json.Marshal(rb)
	if err != nil {
		return nil, nil, err
	}

	bufferedReq := bytes.NewBuffer(rbBytes)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/chat/completions", oai.baseURL), bufferedReq)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", oai.apiKey))
	req.Header.Add("content-type", "application/json")

	responseCh := make(chan Message)
	deltaCh := make(chan string)

	go func() {
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending request to OpenAI: %s", err)
			return
		}

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading contents: %s", err)
			return
		}
		log.Printf("Response: %s", string(respBytes))

		if resp.StatusCode >= 400 {
			log.Printf("Error calling OpenAI %s", string(respBytes))
			return
		}

		var response openAIResponse
		err = json.Unmarshal(respBytes, &response)
		if err != nil {
			log.Printf("Error unmarshalling response: %s", err)
			return
		}
		log.Printf("Response 2: %+v", response)

		msg := Message{
			Role: response.Choices[0].Message.Role,
			Content: []Content{
				{
					Text:        response.Choices[0].Message.Content,
					ContentType: "text",
				},
			},
		}
		log.Printf("msg: %+v", msg)

		deltaCh <- response.Choices[0].Message.Content
		close(deltaCh)

		responseCh <- msg
		close(responseCh)
	}()

	return responseCh, deltaCh, nil
}

type openAIReqBody struct {
	MaxTokens int       `json:"max_completion_tokens"`
	Messages  []Message `json:"messages"`
	Model     string    `json:"model"`
	Stream    bool      `json:"stream,omitempty"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Id      string         `json:"id"`
	Model   string         `json:"model"`
}

type openAIChoice struct {
	Message Content `json:"message"`
}

// type openAIStreamingResponse struct {
// 	Type  string              `json:"type"`
// 	Index int                 `json:"index"`
// 	Delta openAIResponseDelta `json:"delta"`
// }

// type openAIResponseDelta struct {
// 	Type string `json:"type"`
// 	Text string `json:"text"`
// }

// type MessagesEventMessageStartData struct {
// 	Type    string            `json:"type"`
// 	Message anthropicResponse `json:"message"`
// }

// type MessagesEventContentBlockStartData struct {
// 	Type         string  `json:"type"`
// 	Index        int     `json:"index"`
// 	ContentBlock Content `json:"content_block"`
// }

// type MessagesEventContentBlockDeltaData struct {
// 	Type  string  `json:"type"`
// 	Index int     `json:"index"`
// 	Delta Content `json:"delta"`
// }

// type MessagesEventContentBlockStopData struct {
// 	Type  string `json:"type"`
// 	Index int    `json:"index"`
// }

// type MessagesEventMessageDeltaData struct {
// 	Type  string            `json:"type"`
// 	Delta anthropicResponse `json:"delta"`
// 	Usage map[string]int    `json:"usage"`
// }

// type MessagesEventMessageStopData struct {
// 	Type string `json:"type"`
// }
