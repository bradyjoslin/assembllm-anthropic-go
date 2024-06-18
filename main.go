package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/extism/go-pdk"
)

type Model struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Body   RequestBody
	ApiKey string
	Url    string
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

type RequestBody struct {
	Model       string    `json:"model"`
	System      string    `json:"system"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
}

type Content struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type CompletionsResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Model        string    `json:"model"`
	Content      []Content `json:"content"`
	StopReason   string    `json:"stop_reason"`
	StopSequence string    `json:"stop_sequence"`
	Usage        Usage     `json:"usage"`
}

type CompletionToolInput struct {
	Tools    []Tool    `json:"tools"`
	Messages []Message `json:"messages"`
}

type Output struct {
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

var models = []Model{
	{
		Name:    "claude-3-opus-20240229",
		Aliases: []string{"opus"},
	},
	{
		Name:    "claude-3-sonnet-20240229",
		Aliases: []string{"sonet"},
	},
	{
		Name:    "claude-3-haiku-20240307",
		Aliases: []string{"haiku"},
	},
}

var (
	API_VERSION = "2023-06-01"
	MAXTOKENS   = 4096
)

//go:export models
func Models() int32 {
	modelsJson, err := json.Marshal(models)
	if err != nil {
		pdk.OutputString("Error converting models to JSON: " + err.Error())
		return 1
	}

	pdk.Log(pdk.LogInfo, "Returning models")
	pdk.OutputString(string(modelsJson))
	return 0
}

func setTemperature(temperature string) (float64, error) {
	temperatureFloat, err := strconv.ParseFloat(temperature, 32)
	if err != nil {
		return 0, fmt.Errorf("Temperature must be a float: %v", err)
	}
	if temperatureFloat < 0.0 || temperatureFloat > 1.0 {
		return 0, fmt.Errorf("Temperature must be between 0.0 and 1.0")
	}

	return temperatureFloat, nil
}

func setModel(model string) (string, error) {
	var validModel string
	for _, m := range models {
		if model == m.Name {
			validModel = model
			break
		}
		for _, alias := range m.Aliases {
			if model == alias {
				validModel = m.Name
				break
			}
		}
	}
	if validModel == "" {
		return "", fmt.Errorf("Invalid model")
	}

	return validModel, nil
}

func (cReq CompletionRequest) getCompletionsResponse() (CompletionsResponse, error) {
	jsonData, err := json.Marshal(cReq.Body)
	if err != nil {
		fmt.Println(err)
		return CompletionsResponse{}, err
	}

	req := pdk.NewHTTPRequest(pdk.MethodPost, cReq.Url)
	req.SetBody(jsonData)
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("x-api-key", cReq.ApiKey)
	req.SetHeader("anthropic-version", API_VERSION)

	res := req.Send()
	if res.Status() != 200 {
		pdk.Log(pdk.LogError, fmt.Sprintf("Error sending request: %v", res.Status()))
		return CompletionsResponse{}, fmt.Errorf("Error sending request: %v\n%v", res.Status(), string(res.Body()))
	}

	body := res.Body()

	var completionsResponse CompletionsResponse
	err = json.Unmarshal([]byte(body), &completionsResponse)
	if err != nil {
		pdk.Log(pdk.LogError, "Error unmarshalling response: "+err.Error())
		return CompletionsResponse{}, fmt.Errorf("Error unmarshalling response: %v", err)
	}
	return completionsResponse, nil
}

func getConfigValues() (string, string, float64, string, error) {
	api_key, ok := pdk.GetConfig("api_key")
	if !ok {
		return "", "", 0, "", errors.New("api_key empty")
	}

	role, ok := pdk.GetConfig("role")
	if !ok {
		pdk.Log(pdk.LogInfo, "Role not set")
	}

	requested_temperature, _ := pdk.GetConfig("temperature")
	if requested_temperature == "" {
		pdk.Log(pdk.LogInfo, "Temperature not set, using default value")
		requested_temperature = "0.7"
	}

	temperature, err := setTemperature(requested_temperature)
	if err != nil {
		return "", "", 0, "", err
	}

	requested_model, ok := pdk.GetConfig("model")
	if !ok {
		pdk.Log(pdk.LogInfo, "Model not set, using default value")
		requested_model = models[0].Name
	}

	model, err := setModel(requested_model)
	if err != nil {
		return "", "", 0, "", err
	}

	return api_key, role, temperature, model, nil
}

func createCompletionRequest(api_key, role string, temperature float64, model string, messages []Message, tools []Tool) CompletionRequest {
	return CompletionRequest{
		Body: RequestBody{
			Model:       model,
			Temperature: temperature,
			MaxTokens:   MAXTOKENS,
			System:      role,
			Tools:       tools,
			Messages:    messages,
		},
		ApiKey: api_key,
		Url:    "https://api.anthropic.com/v1/messages",
	}
}

//go:export completion
func Completion() int32 {
	prompt := pdk.InputString()

	api_key, role, temperature, model, err := getConfigValues()
	if err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("Error getting config values: %v", err.Error()))
		return 1
	}

	pdk.Log(pdk.LogInfo, "Prompt: "+prompt)

	completionRequest := createCompletionRequest(api_key, role, temperature, model, []Message{{Role: "user", Content: prompt}}, nil)

	completionResponse, err := completionRequest.getCompletionsResponse()
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	pdk.OutputString(completionResponse.Content[0].Text)
	return 0
}

//go:export completionWithTools
func CompletionWithTools() int32 {
	var input CompletionToolInput
	err := pdk.InputJSON(&input)
	if err != nil {
		pdk.Log(pdk.LogError, "Error unmarshalling input: "+err.Error())
		pdk.SetError(fmt.Errorf("Error unmarshalling input: %v", err))
		return 1
	}

	api_key, role, temperature, model, err := getConfigValues()
	if err != nil {
		pdk.Log(pdk.LogError, fmt.Sprintf("Error getting config values: %v", err.Error()))
		return 1
	}

	completionRequest := createCompletionRequest(api_key, role, temperature, model, input.Messages, input.Tools)

	completionResponse, err := completionRequest.getCompletionsResponse()
	if err != nil {
		pdk.SetError(err)
		return 1
	}

	if len(completionResponse.Content) < 2 {
		pdk.SetError(errors.New("no tool response"))
		return 1
	}

	outputs := make([]Output, 0, len(completionResponse.Content)-1)

	for i := 1; i < len(completionResponse.Content); i++ {
		outputs = append(outputs, Output{
			Name:  completionResponse.Content[i].Name,
			Input: completionResponse.Content[i].Input,
		})
	}

	pdk.OutputJSON(outputs)
	return 0
}

func main() {}
