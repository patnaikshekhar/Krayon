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
)

type openai struct {
	apiKey  string
	baseURL string
	stream  bool
}

func NewOpenAI(apiKey string, stream bool) *openai {
	return &openai{apiKey, "https://api.openai.com/v1", stream}
}

func (oai *openai) Chat(ctx context.Context, model string, temperature int32, messages []Message, tools []Tool) (<-chan Message, <-chan string, error) {

	// Map to OpenAI Format
	var oaiMessages []OpenAIMessage
	for _, m := range messages {
		if m.Role == "plugin" {
			continue
		}

		var oaiContent []OpenAIContent
		for _, c := range m.Content {
			if c.ContentType == "text" {
				oaiContent = append(oaiContent, OpenAIContent{
					Type: "text",
					Text: c.Text,
				})
			} else if c.ContentType == "image" {
				oaiContent = append(oaiContent, OpenAIContent{
					Type: "image_url",
					ImageUrl: &OpenAIImageUrl{
						Url: fmt.Sprintf("data:%s;base64,%s", c.Source.MediaType, c.Source.Data),
					},
				})
			}
		}
		oaiMessages = append(oaiMessages, OpenAIMessage{
			Role:    m.Role,
			Content: oaiContent,
		})
	}

	rb := openAIReqBody{
		Messages:  oaiMessages,
		MaxTokens: 4096,
		Model:     model,
		Stream:    oai.stream,
	}

	rbBytes, err := json.Marshal(rb)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("Request: %s", string(rbBytes))

	bufferedReq := bytes.NewBuffer(rbBytes)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/chat/completions", oai.baseURL), bufferedReq)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", oai.apiKey))
	req.Header.Add("content-type", "application/json")

	responseCh := make(chan Message)
	deltaCh := make(chan string)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request to OpenAI: %s", err)
		return nil, nil, err
	}

	if resp.StatusCode >= 400 {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading contents: %s", err)
			return nil, nil, errors.New(resp.Status)
		}
		log.Printf("Error calling OpenAI %s", string(respBytes))
		return nil, nil, errors.New(string(respBytes))
	}

	if !oai.stream {
		go func() {
			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading contents: %s", err)
				return
			}

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

			msg := Message{
				Role: response.Choices[0].Message.Role,
				Content: []Content{
					{
						Text:        response.Choices[0].Message.Content,
						ContentType: "text",
					},
				},
			}

			deltaCh <- response.Choices[0].Message.Content
			close(deltaCh)

			responseCh <- msg
			close(responseCh)
		}()
	} else {
		go func() {
			defer close(responseCh)

			dataPrefix := []byte("data: ")
			response := openAIResponse{
				Choices: []openAIChoice{
					{
						Message: Content{
							ContentType: "text",
							Text:        "",
						},
					},
				},
			}
			reader := bufio.NewReader(resp.Body)
			for {
				rawLine, readErr := reader.ReadBytes('\n')
				if readErr != nil {
					if errors.Is(readErr, io.EOF) {
						break
					}
					log.Printf("Error reading from OpenAI: %s", readErr)
					break
				}

				noSpaceLine := bytes.TrimSpace(rawLine)
				if len(noSpaceLine) == 0 {
					continue
				}

				if bytes.HasPrefix(noSpaceLine, dataPrefix) {
					data := bytes.TrimSpace(bytes.TrimPrefix(noSpaceLine, dataPrefix))
					if len(data) == 0 {
						continue
					}

					if string(data) == "[DONE]" {
						break
					}

					sr := openAIStreamingResponse{}
					err = json.Unmarshal(data, &sr)
					if err != nil {
						log.Printf("Error unmarshalling streaming response %s", err)
						continue
					}

					deltaCh <- sr.StreamingChoices[0].Delta.Content
					response.Choices[0].Message.Content += sr.StreamingChoices[0].Delta.Content
				}
			}

			close(deltaCh)

			msg := Message{
				Role:    "assistant",
				Content: []Content{response.Choices[0].Message},
			}

			responseCh <- msg
		}()
	}

	return responseCh, deltaCh, nil
}

type openAIReqBody struct {
	MaxTokens int             `json:"max_completion_tokens"`
	Messages  []OpenAIMessage `json:"messages"`
	Model     string          `json:"model"`
	Stream    bool            `json:"stream,omitempty"`
}

type OpenAIMessage struct {
	Role    string          `json:"role"`
	Content []OpenAIContent `json:"content"`
}

type OpenAIContent struct {
	Type     string          `json:"type,omitempty"`
	Text     string          `json:"text,omitempty"`
	ImageUrl *OpenAIImageUrl `json:"image_url,omitempty"`
}

type OpenAIImageUrl struct {
	Url string `json:"url"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Id      string         `json:"id"`
	Model   string         `json:"model"`
}

type openAIStreamingResponse struct {
	Id               string                  `json:"id"`
	Model            string                  `json:"model"`
	StreamingChoices []streamingOpenAIChoice `json:"choices"`
}

type streamingOpenAIChoice struct {
	Index int     `json:"index"`
	Delta Content `json:"delta"`
}
type openAIChoice struct {
	Message Content `json:"message"`
}
