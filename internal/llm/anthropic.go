package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
)

type anthropic struct {
	apiKey  string
	baseURL string
}

func NewAnthropic(apiKey string) *anthropic {
	return &anthropic{apiKey, "https://api.anthropic.com/v1"}
}

func (a *anthropic) Chat(ctx context.Context, model string, temperature int32, messages []Message, tools []Tool) (<-chan Message, <-chan string, error) {
	systemMessage := ""
	var cleansedMessages []Message
	for _, sm := range messages {
		if sm.Role == "system" {
			systemMessage += sm.Content[0].Text
			continue
		}

		if sm.Role == "plugin" {
			continue
		}

		cleansedMessages = append(cleansedMessages, sm)
	}

	rb := anthropicReqBody{
		MaxTokens: 4096,
		Model:     model,
		Messages:  cleansedMessages,
		System:    systemMessage,
		Tools:     tools,
		Stream:    true,
	}

	rbBytes, err := json.Marshal(rb)
	if err != nil {
		return nil, nil, err
	}

	bufferedReq := bytes.NewBuffer(rbBytes)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/messages", a.baseURL), bufferedReq)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Add("x-api-key", a.apiKey)
	req.Header.Add("anthropic-version", "2023-06-01")
	req.Header.Add("content-type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode >= 400 {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("Error calling Anthropic %s", string(respBytes))
	}

	responseCh := make(chan Message)
	deltaCh := make(chan string)

	go func() {
		defer close(responseCh)
		var event []byte
		eventPrefix := []byte("event: ")
		dataPrefix := []byte("data: ")
		response := anthropicResponse{}
		reader := bufio.NewReader(resp.Body)
		for {
			rawLine, readErr := reader.ReadBytes('\n')
			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					break
				}
				log.Printf("Error reading from Anthropic: %s", readErr)
				break
			}

			noSpaceLine := bytes.TrimSpace(rawLine)
			if len(noSpaceLine) == 0 {
				continue
			}

			if bytes.HasPrefix(noSpaceLine, eventPrefix) {
				event = bytes.TrimSpace(bytes.TrimPrefix(noSpaceLine, eventPrefix))
				continue
			}

			if bytes.HasPrefix(noSpaceLine, dataPrefix) {
				data := bytes.TrimSpace(bytes.TrimPrefix(noSpaceLine, dataPrefix))
				if len(data) == 0 {
					continue
				}
				// log.Printf("Event is %s", event)
				// log.Printf("Data is %s", data)

				sr := anthropicStreamingResponse{}
				err = json.Unmarshal(data, &sr)
				if err != nil {
					log.Printf("Error unmarshalling streaming response %s", err)
					continue
				}

				if sr.Type == "content_block_delta" && sr.Delta.Type == "text_delta" {
					deltaCh <- sr.Delta.Text
				}

				switch string(event) {
				case "message_start":
					var d MessagesEventMessageStartData
					if err := json.Unmarshal(data, &d); err != nil {
						log.Printf("Error unmarshalling streaming response %s", err)
						continue
					}
					response = d.Message
					continue
				case "content_block_start":
					var d MessagesEventContentBlockStartData
					if err := json.Unmarshal(data, &d); err != nil {
						log.Printf("Error unmarshalling streaming response %s", err)
						continue
					}

					response.Content = slices.Insert(response.Content, d.Index, d.ContentBlock)
				case "content_block_delta":
					var d MessagesEventContentBlockDeltaData
					if err := json.Unmarshal(data, &d); err != nil {
						log.Printf("Error unmarshalling streaming response %s", err)
						continue
					}
					if len(response.Content)-1 < d.Index {
						response.Content = slices.Insert(response.Content, d.Index, d.Delta)
					} else {
						response.Content[d.Index].MergeContentDelta(d.Delta)
					}
				case "content_block_stop":
					var d MessagesEventContentBlockStopData
					if err := json.Unmarshal(data, &d); err != nil {
						log.Printf("Error unmarshalling streaming response %s", err)
						continue
					}
					var stopContent Content
					if len(response.Content) > d.Index {
						stopContent = response.Content[d.Index]
						if stopContent.ContentType == "tool_use" {
							var res map[string]interface{}
							err = json.Unmarshal([]byte(stopContent.Content), &res)
							if err != nil {
								log.Printf("Error unmarshalling partial json response %s", err)
								continue
							}
							stopContent.Input = res
							stopContent.PartialJson = nil
							response.Content[d.Index] = stopContent
						}
					}
					continue
				case "message_delta":
					var d MessagesEventMessageDeltaData
					if err := json.Unmarshal(data, &d); err != nil {
						log.Printf("Error unmarshalling streaming response %s", err)
						continue
					}
					response.Usage = d.Usage
					continue
				}
			}
		}

		close(deltaCh)

		msg := Message{
			Role:    "assistant",
			Content: response.Content,
		}

		responseCh <- msg
	}()

	return responseCh, deltaCh, nil
}

type anthropicReqBody struct {
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Model     string    `json:"model"`
	System    string    `json:"system"`
	Tools     []Tool    `json:"tools,omitempty"`
	Stream    bool      `json:"stream,omitempty"`
}

type anthropicResponse struct {
	Content []Content      `json:"content"`
	Id      string         `json:"id"`
	Model   string         `json:"model"`
	Usage   map[string]int `json:"usage"`
}

type anthropicStreamingResponse struct {
	Type  string                 `json:"type"`
	Index int                    `json:"index"`
	Delta anthropicResponseDelta `json:"delta"`
}

type anthropicResponseDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type MessagesEventMessageStartData struct {
	Type    string            `json:"type"`
	Message anthropicResponse `json:"message"`
}

type MessagesEventContentBlockStartData struct {
	Type         string  `json:"type"`
	Index        int     `json:"index"`
	ContentBlock Content `json:"content_block"`
}

type MessagesEventContentBlockDeltaData struct {
	Type  string  `json:"type"`
	Index int     `json:"index"`
	Delta Content `json:"delta"`
}

type MessagesEventContentBlockStopData struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

type MessagesEventMessageDeltaData struct {
	Type  string            `json:"type"`
	Delta anthropicResponse `json:"delta"`
	Usage map[string]int    `json:"usage"`
}

type MessagesEventMessageStopData struct {
	Type string `json:"type"`
}
