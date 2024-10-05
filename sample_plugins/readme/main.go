package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"krayon/internal/llm"
	"krayon/internal/plugins"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

//go:embed README.md
var readmeContents string

func main() {
	f, err := os.OpenFile("readme.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	socket := flag.String("socket", "", "Unix socket to listen on")
	flag.Parse()
	if socket == nil || *socket == "" {
		fmt.Println("No socket provided")
		os.Exit(-1)
		return
	}

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", *socket)
			},
		},
	}

	time.Sleep(5 * time.Second)

	response, err := client.Get("http://unix/api/v1/connect")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
		return
	}

	var info plugins.RequestInfo
	err = json.NewDecoder(response.Body).Decode(&info)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
		return
	}

	go sendResponse(client, plugins.PluginResponse{
		Data:  "Sending request to generate a README\n",
		Error: "",
	})

	log.Printf("Constructing llmRequestBody")

	llmRequestBody := plugins.LlmRequest{
		Messages: []llm.Message{
			{
				Role: "system",
				Content: []llm.Content{
					{
						Text: `You are an experienced engineer who is excellent at tech documentation. 
										Could you review the provided README and generate an exemplar README from it for the project? `,
						ContentType: "text",
					},
				},
			},
			{
				Role: "user",
				Content: []llm.Content{
					{
						Text:        string(readmeContents),
						ContentType: "text",
					},
				},
			},
		},
	}

	log.Printf("llmRequestBody is %+v", llmRequestBody)

	buf := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(buf).Encode(llmRequestBody)
	if err != nil {
		log.Printf("Error making a request to the llm %s", err)
		os.Exit(-1)
		return
	}

	log.Printf("Calling /llm")
	resp, err := client.Post("http://unix/api/v1/llm", "application/json", buf)
	log.Printf("Called /llm")
	if err != nil {
		log.Printf("Error in response from the /llm endpoint %s", err)
		os.Exit(-1)
		return
	}

	llmResponse := llm.Message{}
	err = json.NewDecoder(resp.Body).Decode(&llmResponse)
	log.Printf("LLM response %+v", &llmResponse)
	if err != nil {
		log.Printf("Error reading response from the LLM %s", err)
		os.Exit(-1)
		return
	}

	go sendResponse(client, plugins.PluginResponse{
		Data:  "\nWriting README",
		Error: "",
	})

	log.Printf("Writing README")
	err = os.WriteFile("README-AI.md", []byte(llmResponse.Content[0].Text), os.ModePerm)
	if err != nil {

		go sendResponse(client, plugins.PluginResponse{
			Data:  "",
			Error: err.Error(),
		})
		os.Exit(-1)
		return
	}

}

func sendResponse(client http.Client, response plugins.PluginResponse) {
	buf := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buf).Encode(response)
	if err != nil {
		log.Printf("Error sending response to /response %s", err)
		os.Exit(-1)
		return
	}

	_, err = client.Post("http://unix/api/v1/response", "application/json", buf)
	if err != nil {
		log.Printf("Error in response from the /response endpoint %s", err)
		os.Exit(-1)
		return
	}
}
