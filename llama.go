package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type LlamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context"`
	TotalDuration      int    `json:"total_duration"`
	LoadDuration       int    `json:"load_duration"`
	PromptEvalDuration int    `json:"prompt_eval_duration"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int    `json:"eval_duration"`
}

const url string = "http://localhost:11434/api/generate"

func StreamRequest(prompt string, dataChannel chan<- string) {
	jsonBody := []byte(`{
		"model": "llama3:8b",
		"prompt": "` + prompt + `"
	}`)

	bodyReader := bytes.NewReader(jsonBody)

	resp, err := http.Post(url, "application/json", bodyReader)

	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	defer resp.Body.Close()

	var response LlamaResponse

	json.NewDecoder(resp.Body).Decode(&response)
	dataChannel <- response.Response
	for !response.Done {
		json.NewDecoder(resp.Body).Decode(&response)
		dataChannel <- response.Response
	}
	close(dataChannel)
}
