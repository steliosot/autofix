package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

type Suggestion struct {
	Explanation string    `json:"explanation"`
	ProposedFix string    `json:"proposed_fix"`
	RiskLevel   RiskLevel `json:"risk_level"`
}

type Environment struct {
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	Architecture   string `json:"architecture"`
	PackageManager string `json:"package_manager"`
	HasSudo        bool   `json:"has_sudo"`
	InContainer    bool   `json:"in_container"`
}

type Request struct {
	Environment Environment `json:"environment"`
	Command     string      `json:"command"`
	Stderr      string      `json:"stderr"`
	ExitCode    int         `json:"exit_code"`
	Attempt     int         `json:"attempt"`
}

type Client interface {
	GetSuggestion(req *Request) (*Suggestion, error)
}

type MockClient struct{}

func (m *MockClient) GetSuggestion(req *Request) (*Suggestion, error) {
	return &Suggestion{
		Explanation: "Mock suggestion: This is a placeholder for LLM response",
		ProposedFix: "echo 'Mock fix applied'",
		RiskLevel:   RiskLow,
	}, nil
}

type OpenAIClient struct {
	APIKey   string
	Endpoint string
	Model    string
}

func (o *OpenAIClient) GetSuggestion(req *Request) (*Suggestion, error) {
	if o.APIKey == "" {
		apiKey := os.Getenv("OPENAI_API_KEY")
		o.APIKey = apiKey
	}

	body, _ := json.Marshal(map[string]interface{}{
		"model": o.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful DevOps assistant. Respond with JSON containing explanation, proposed_fix, and risk_level."},
			{"role": "user", "content": fmt.Sprintf("Command: %s\nError: %s\nEnvironment: %+v", req.Command, req.Stderr, req.Environment)},
		},
		"response_format": map[string]string{"type": "json_object"},
	})

	httpReq, _ := http.NewRequest("POST", o.Endpoint+"/chat/completions", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.APIKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(respBody, &result)

	if len(result.Choices) == 0 {
		return &Suggestion{}, nil
	}

	var sug Suggestion
	json.Unmarshal([]byte(result.Choices[0].Message.Content), &sug)
	return &sug, nil
}

func NewClient(provider, apiKey, endpoint, model string) Client {
	switch provider {
	case "mock":
		return &MockClient{}
	case "openai":
		return &OpenAIClient{APIKey: apiKey, Endpoint: endpoint, Model: model}
	default:
		if provider == "local" {
			return &MockClient{}
		}
		return &OpenAIClient{APIKey: apiKey, Endpoint: endpoint, Model: model}
	}
}
