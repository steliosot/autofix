package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/autofix/cli/internal/config"
	"github.com/autofix/cli/internal/env"
	"github.com/autofix/cli/internal/fixengine"
	"github.com/autofix/cli/internal/llm"
	"github.com/autofix/cli/internal/safety"
)

const Version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	config.Init()

	command := os.Args[1]

	switch command {
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: command required")
			printUsage()
			os.Exit(1)
		}
		runCommand(strings.Join(os.Args[2:], " "))
	case "config":
		if len(os.Args) < 4 {
			fmt.Println("Error: config command requires key and value")
			fmt.Println("Usage: autofix config <key> <value>")
			os.Exit(1)
		}
		if err := config.Set(os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration updated")
	case "setup":
		runSetup()
	case "version":
		fmt.Printf("AutoFix %s\n", Version)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("AutoFix - Self-healing DevOps Assistant")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  autofix run <command>    Execute command with auto-healing")
	fmt.Println("  autofix config <key> <value>  Set configuration")
	fmt.Println("  autofix setup           Interactive setup")
	fmt.Println("  autofix version         Show version")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  autofix run 'npm install'")
	fmt.Println("  autofix run 'pip install requests'")
	fmt.Println("  autofix config set llm.api_key sk-...")
}

func runCommand(cmd string) {
	fmt.Println("[Detecting Environment]")
	environment := env.Detect()
	fmt.Printf("OS: %s\n", environment.OS)
	fmt.Printf("Architecture: %s\n", environment.Architecture)
	fmt.Printf("Package Manager: %s\n", environment.PackageManager)
	fmt.Printf("Has Sudo: %v\n", environment.HasSudo)
	fmt.Printf("In Container: %v\n", environment.InContainer)

	cfg := config.Get()

	llmClient := llm.NewClient(cfg.LLM.Provider, cfg.LLM.APIKey, cfg.LLM.Endpoint, cfg.LLM.Model)

	fixEngine := fixengine.New(environment, llmClient)

	validator := safety.NewValidator()
	if err := validator.Validate(cmd); err != nil {
		fmt.Printf("[Safety Check] %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[Executing Command]")
	fmt.Printf("Command: %s\n", cmd)

	result, err := fixEngine.ExecuteWithRetry(cmd, 0)
	if err != nil {
		fmt.Printf("[Error] %v\n", err)
		os.Exit(1)
	}

	if result.Success {
		fmt.Println("[Success]")
	} else {
		fmt.Println("[Failed]")
		os.Exit(result.ExitCode)
	}
}

func runSetup() {
	fmt.Println("AutoFix Setup")
	fmt.Println("===============")
	fmt.Println()

	var provider string
	fmt.Print("LLM Provider (openai/local): ")
	fmt.Scanln(&provider)
	if provider == "" {
		provider = "openai"
	}

	var apiKey string
	if provider != "mock" {
		fmt.Print("API Key (leave empty for mock): ")
		fmt.Scanln(&apiKey)
	}

	var endpoint string
	fmt.Print("API Endpoint (default: https://api.openai.com/v1): ")
	fmt.Scanln(&endpoint)
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	var model string
	fmt.Print("Model (default: gpt-4): ")
	fmt.Scanln(&model)
	if model == "" {
		model = "gpt-4"
	}

	var autoExecute string
	fmt.Print("Auto-execute low-risk fixes? (y/N): ")
	fmt.Scanln(&autoExecute)

	config.Set("llm.provider", provider)
	config.Set("llm.api_key", apiKey)
	config.Set("llm.endpoint", endpoint)
	config.Set("llm.model", model)
	config.Set("safety.auto_execute", autoExecute)

	fmt.Println()
	fmt.Println("Setup complete!")
	fmt.Println("Configuration saved to ~/.autofix/config.yaml")
}
