package plugins

import (
	"encoding/json"
	"krayon/internal/config"
	"krayon/internal/llm"
	"log"
	"net"
	"net/http"
)

type RequestInfo struct {
	Question string
	History  []llm.Message
	Context  string
}

type PluginResponse struct {
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type LlmRequest struct {
	Messages []llm.Message `json:"messages"`
}

func RunPluginServer(requestInfo chan RequestInfo, provider llm.Provider, profile *config.Profile, socketPath string, pluginResponseCh chan string) error {

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/connect", func(w http.ResponseWriter, r *http.Request) {
		ri := <-requestInfo

		err := json.NewEncoder(w).Encode(ri)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/api/v1/response", func(w http.ResponseWriter, r *http.Request) {

		pr := PluginResponse{}
		err := json.NewDecoder(r.Body).Decode(&pr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if pr.Error != "" {
			pluginResponseCh <- pr.Error
			return
		}

		pluginResponseCh <- pr.Data

	})

	mux.HandleFunc("/api/v1/llm", func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		llmReq := LlmRequest{}
		err := json.NewDecoder(r.Body).Decode(&llmReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		msgCh, streamCh, err := provider.Chat(r.Context(), profile.Model, 0, llmReq.Messages, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _ = range streamCh {
		}

		finalMessage := <-msgCh

		err = json.NewEncoder(w).Encode(finalMessage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	})

	server := http.Server{
		Handler: mux,
	}

	log.Printf("Starting server on %s", socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	err = server.Serve(listener)
	if err != nil {
		return err
	}

	return nil
}
