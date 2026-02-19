package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type LLMProvider string

const (
	ProviderOpenAI LLMProvider = "openai"
	ProviderLocal  LLMProvider = "local"
	ProviderHybrid LLMProvider = "hybrid"
)

type RiskLevel string

const (
	RiskAuto    RiskLevel = "auto"
	RiskConfirm RiskLevel = "confirm"
	RiskManual  RiskLevel = "manual"
)

type Config struct {
	LLM struct {
		Provider string `yaml:"provider"`
		APIKey   string `yaml:"api_key"`
		Endpoint string `yaml:"endpoint"`
		Model    string `yaml:"model"`
	} `yaml:"llm"`
	Safety struct {
		AutoExecute        bool `yaml:"auto_execute"`
		RequireSudoConfirm bool `yaml:"require_sudo_confirm"`
	} `yaml:"safety"`
}

var (
	cfg        *Config
	configPath string
)

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath = filepath.Join(homeDir, ".autofix", "config.yaml")

	cfg = &Config{}
	cfg.LLM.Provider = "openai"
	cfg.LLM.Endpoint = "https://api.openai.com/v1"
	cfg.LLM.Model = "gpt-4"
	cfg.Safety.RequireSudoConfirm = true

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		Save()
	}

	return nil
}

func Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".autofix")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func Get() *Config {
	return cfg
}

func Set(key, value string) error {
	switch key {
	case "llm.provider":
		cfg.LLM.Provider = value
	case "llm.api_key":
		cfg.LLM.APIKey = value
	case "llm.endpoint":
		cfg.LLM.Endpoint = value
	case "llm.model":
		cfg.LLM.Model = value
	case "safety.auto_execute":
		cfg.Safety.AutoExecute = (value == "true")
	case "safety.require_sudo_confirm":
		cfg.Safety.RequireSudoConfirm = (value == "true")
	}
	return Save()
}
