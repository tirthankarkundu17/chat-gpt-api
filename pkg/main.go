package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type TextCompletionRequest struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	Temperature      float64 `json:"temperature"`
	MaxTokens        int     `json:"max_tokens"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
}

type Choice struct {
	Text         string      `json:"text"`
	Index        int         `json:"index"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

type TextCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type APIResponse struct {
	Status string    `json:"status"`
	Data   *[]Choice `json:"data"`
}

type APIRequest struct {
	Prompt string `json:"prompt"`
}

var (
	// ChatGPTHTTPAddress  Address
	ChatGPTHTTPAddress = "https://api.openai.com/v1/completions"

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("non 200 response found")

	// Bearer Token
	BearerToken = os.Getenv("CHAT_GPT_TOKEN")
)

func ConverseWithGPT(prompt string) (TextCompletionResponse, error) {
	httpReq := TextCompletionRequest{
		Model:            "text-davinci-003",
		Prompt:           prompt,
		Temperature:      0.7,
		MaxTokens:        100,
		TopP:             1.0,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
	}
	jsonValue, _ := json.Marshal(httpReq)

	bearer := "Bearer " + BearerToken

	req, _ := http.NewRequest("POST", ChatGPTHTTPAddress, bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return TextCompletionResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return TextCompletionResponse{}, ErrNon200Response
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TextCompletionResponse{}, err
	}

	var textCompletionResponse TextCompletionResponse
	if err := json.Unmarshal(body, &textCompletionResponse); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}

	return textCompletionResponse, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var apiRequest APIRequest
	if err := json.Unmarshal([]byte(request.Body), &apiRequest); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	textCompletionResponse, err := ConverseWithGPT(apiRequest.Prompt)

	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	var choices = textCompletionResponse.Choices
	response := APIResponse{
		Status: "Success",
		Data:   &choices,
	}

	jsonBytes, _ := json.Marshal(response)
	rstr := fmt.Sprint(string(jsonBytes))

	return events.APIGatewayProxyResponse{
		Body:       rstr,
		StatusCode: 200,
	}, nil
}

func PrettyPrint(textCompletionResponse TextCompletionResponse) {
	panic("unimplemented")
}

func main() {
	lambda.Start(handler)
}
