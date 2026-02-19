package fixengine

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/autofix/cli/internal/config"
	"github.com/autofix/cli/internal/env"
	"github.com/autofix/cli/internal/errorparser"
	"github.com/autofix/cli/internal/executor"
	"github.com/autofix/cli/internal/llm"
)

const MaxRetries = 3

type FixEngine struct {
	Environment *env.Environment
	LLMClient   llm.Client
}

func New(e *env.Environment, llmClient llm.Client) *FixEngine {
	return &FixEngine{
		Environment: e,
		LLMClient:   llmClient,
	}
}

func (f *FixEngine) ExecuteWithRetry(command string, attempt int) (*executor.Result, error) {
	args := strings.Fields(command)
	if len(args) == 0 {
		args = []string{"/bin/sh", "-c", command}
	}

	cmd := exec.Command(args[0], args[1:]...)
	result, err := executor.Runner(cmd)

	if result.Success {
		return result, nil
	}

	if attempt >= MaxRetries {
		return result, fmt.Errorf("max retries exceeded")
	}

	errorInfo := errorparser.Parse(result.Stderr, result.ExitCode)
	fixCmd, fixType, err := f.GetFix(errorInfo, command, result.Stderr, attempt)

	if err != nil {
		return result, err
	}

	if fixCmd == "" {
		return result, fmt.Errorf("no fix available")
	}

	fmt.Printf("[Applying Fix] %s\n", fixCmd)

	confirmed := true
	cfg := config.Get()

	if !cfg.Safety.AutoExecute {
		fmt.Print("Execute this fix? (y/N): ")
		var response string
		fmt.Scanln(&response)
		confirmed = (strings.ToLower(response) == "y")
	}

	if !confirmed {
		return result, fmt.Errorf("fix declined by user")
	}

	if isSudoCommand(fixCmd) && cfg.Safety.RequireSudoConfirm {
		fmt.Print("This command requires sudo. Execute? (y/N): ")
		var response string
		fmt.Scanln(&response)
		confirmed = (strings.ToLower(response) == "y")
		if !confirmed {
			return result, fmt.Errorf("sudo command declined")
		}
	}

	fixArgs := strings.Fields(fixCmd)
	fixExec := exec.Command(fixArgs[0], fixArgs[1:]...)
	fixResult, _ := executor.Runner(fixExec)

	if !fixResult.Success {
		fmt.Printf("[Fix Failed] %s\n", fixResult.Stderr)
		return result, fmt.Errorf("fix command failed: %s", fixResult.Stderr)
	}

	if fixType == "replacement" {
		fmt.Println("[Success]")
		return fixResult, nil
	}

	fmt.Printf("[Retry %d/%d]\n", attempt+1, MaxRetries)
	return f.ExecuteWithRetry(command, attempt+1)
}

func (f *FixEngine) GetFix(errorInfo *errorparser.ErrorInfo, originalCommand, stderr string, attempt int) (string, string, error) {
	deterministicFix := f.getDeterministicFix(errorInfo)
	if deterministicFix != "" {
		return deterministicFix, "preparation", nil
	}

	if attempt > 0 {
		return "", "", nil
	}

	llmReq := &llm.Request{
		Environment: llm.Environment{
			OS:             string(f.Environment.OS),
			OSVersion:      f.Environment.OSVersion,
			Architecture:   string(f.Environment.Architecture),
			PackageManager: string(f.Environment.PackageManager),
			HasSudo:        f.Environment.HasSudo,
			InContainer:    f.Environment.InContainer,
		},
		Command:  originalCommand,
		Stderr:   stderr,
		ExitCode: -1,
		Attempt:  attempt,
	}

	suggestion, err := f.LLMClient.GetSuggestion(llmReq)
	if err != nil {
		return "", "", err
	}

	fixType := suggestion.FixType
	if fixType == "" {
		fixType = "preparation"
	}

	if suggestion.RiskLevel == llm.RiskLow {
		return suggestion.ProposedFix, fixType, nil
	}

	cfg := config.Get()
	if cfg.Safety.AutoExecute {
		return suggestion.ProposedFix, fixType, nil
	}

	fmt.Printf("[LLM Suggestion] %s\n", suggestion.Explanation)
	fmt.Printf("[Proposed Fix] %s\n", suggestion.ProposedFix)
	fmt.Printf("[Risk Level] %s\n", suggestion.RiskLevel)

	fmt.Print("Apply this fix? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" {
		return suggestion.ProposedFix, fixType, nil
	}

	return "", "", nil
}

func (f *FixEngine) getDeterministicFix(errorInfo *errorparser.ErrorInfo) string {
	switch errorInfo.Type {
	case errorparser.ErrorTypeMissingCommand:
		return f.installPackage(errorInfo.Command)
	case errorparser.ErrorTypeMissingCompiler:
		return f.installBuildEssential()
	case errorparser.ErrorTypeMissingLibrary:
		return f.installPackage(errorInfo.Package)
	case errorparser.ErrorTypeMissingBuildTools:
		return f.installBuildEssential()
	default:
		return ""
	}
}

func (f *FixEngine) installPackage(pkg string) string {
	if pkg == "" {
		return ""
	}

	switch f.Environment.PackageManager {
	case env.PMApt:
		return fmt.Sprintf("sudo apt-get install -y %s", pkg)
	case env.PMDnf, env.PMYum:
		return fmt.Sprintf("sudo dnf install -y %s", pkg)
	case env.PMPacman:
		return fmt.Sprintf("sudo pacman -S --noconfirm %s", pkg)
	case env.PMBrew:
		return fmt.Sprintf("brew install %s", pkg)
	default:
		return ""
	}
}

func (f *FixEngine) installBuildEssential() string {
	switch f.Environment.PackageManager {
	case env.PMApt:
		return "sudo apt-get install -y build-essential"
	case env.PMDnf:
		return "sudo dnf groupinstall -y 'Development Tools'"
	case env.PMPacman:
		return "sudo pacman -S --noconfirm base-devel"
	case env.PMBrew:
		return "xcode-select --install"
	default:
		return ""
	}
}

func isSudoCommand(cmd string) bool {
	return strings.HasPrefix(cmd, "sudo ") || strings.HasPrefix(cmd, "sudo\t")
}
