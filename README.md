# AutoFix

A production-ready CLI tool for language-agnostic, environment-aware, self-healing DevOps automation. AutoFix executes shell commands, detects failures, and applies intelligent fixes using a deterministic-first approach with optional LLM integration.

## Features

- **Cross-platform static binaries** for Linux (amd64/arm64) and macOS (amd64/arm64)
- **Zero runtime dependencies** - single binary with no external packages
- **Environment-aware detection**: OS, version, architecture, package manager, runtimes, sudo availability
- **Deterministic-first error classification** before LLM calls
- **Multi-provider LLM support**: OpenAI, local endpoints, hybrid mode
- **Safety constraints**: Allowlist validation, sudo confirmation, risk assessment
- **Secure configuration**: Stored in `~/.autofix/config.yaml` with permissions 600

## Installation

### Via Install Script (Non-interactive)

```bash
curl -fsSL https://github.com/steliosot/autofix/raw/main/install.sh | bash
```

### Manually

Download the appropriate binary for your platform:

- Linux AMD64: `autofix-linux-amd64`
- Linux ARM64: `autofix-linux-arm64`
- macOS AMD64: `autofix-darwin-amd64`
- macOS ARM64: `autofix-darwin-arm64`

Move to `/usr/local/bin/autofix` and make executable.

## Usage

```bash
autofix run "npm install"
autofix run "pip install requests"
autofix run "docker build -t myapp ."
autofix config llm.provider openai
autofix config llm.api_key sk-...
autofix setup
autofix version
```

## Architecture

```
cmd/
  autofix/main.go           # Entry point
internal/
  env/
    types.go                # Environment types
    detect.go              # OS/platform detection
  executor/
    executor.go            # Command execution
  errorparser/
    errorparser.go         # Error classification
  fixengine/
    fixengine.go           # Fix application + retry logic
  llm/
    llm.go                # LLM provider interface
  config/
    config.go             # Configuration management
  safety/
    safety.go             # Command validation
```

## Deterministic Fix Rules

| Error Type | Fix Strategy |
|------------|--------------|
| Missing Command | Install via package manager |
| Missing Compiler | Install build-essential / xcode-select |
| Missing Library | Install via package manager |
| Port in Use | Kill process or use different port |
| Permission Denied | Sudo prefix or file permissions |
| Build Tools Missing | Install build toolchain |

## Safety

- Auto-execute only low-risk commands
- Require confirmation for medium/high risk
- Block destructive commands (rm -rf, userdel, etc.)
- Sudo commands always require confirmation unless configured

## Development

```bash
# Build for current platform
make build-local

# Build all platforms
make build

# Clean
make clean

# Test
make test
```

## Configuration

Located at `~/.autofix/config.yaml`:

```yaml
llm:
  provider: openai
  api_key: ""
  endpoint: https://api.openai.com/v1
  model: gpt-4
safety:
  auto_execute: false
  require_sudo_confirm: true
```