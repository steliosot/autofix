package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
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
	apiKey := o.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	body, _ := json.Marshal(map[string]interface{}{
		"model": o.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a DevOps fix assistant. Given a failed command, respond with ONLY a JSON object (no markdown, no explanation outside JSON): {\"explanation\": \"one sentence\", \"proposed_fix\": \"single shell command\", \"risk_level\": \"low\"}. If no fix exists, return {\"explanation\": \"cannot fix\", \"proposed_fix\": \"\", \"risk_level\": \"high\"}"},
			{"role": "user", "content": fmt.Sprintf("Failed command: %s\nStderr: %s\nOS: %s, Package Manager: %s\n\nReturn JSON with proposed_fix as a single shell command or empty if unfixable.", req.Command, req.Stderr, req.Environment.OS, req.Environment.PackageManager)},
		},
	})

	httpReq, _ := http.NewRequest("POST", o.Endpoint+"/chat/completions", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &Suggestion{Explanation: "LLM API call failed", ProposedFix: "", RiskLevel: RiskHigh}, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	json.Unmarshal(respBody, &result)

	if result.Error.Message != "" {
		return &Suggestion{Explanation: "LLM error: " + result.Error.Message, ProposedFix: "", RiskLevel: RiskHigh}, nil
	}

	if len(result.Choices) == 0 {
		return &Suggestion{Explanation: "No LLM response", ProposedFix: "", RiskLevel: RiskHigh}, nil
	}

	content := result.Choices[0].Message.Content
	content = extractJSON(content)

	var sug Suggestion
	json.Unmarshal([]byte(content), &sug)
	return &sug, nil
}

func extractJSON(s string) string {
	if idx := indexOf(s, "```json"); idx >= 0 {
		s = s[idx+7:]
		if end := indexOf(s, "```"); end >= 0 {
			s = s[:end]
		}
	} else if idx := indexOf(s, "```"); idx >= 0 {
		s = s[idx+3:]
		if end := indexOf(s, "```"); end >= 0 {
			s = s[:end]
		}
	}
	start := indexOf(s, "{")
	if start >= 0 {
		s = s[start:]
		depth := 0
		for i, c := range s {
			if c == '{' {
				depth++
			} else if c == '}' {
				depth--
				if depth == 0 {
					return s[:i+1]
				}
			}
		}
	}
	return s
}

func indexOf(s string, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
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
