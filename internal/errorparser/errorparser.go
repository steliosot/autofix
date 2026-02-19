package errorparser

import (
	"strings"
)

type ErrorType string

const (
	ErrorTypeMissingCommand         ErrorType = "missing_command"
	ErrorTypeMissingCompiler        ErrorType = "missing_compiler"
	ErrorTypeMissingLibrary         ErrorType = "missing_library"
	ErrorTypePortInUse              ErrorType = "port_in_use"
	ErrorTypePermissionDenied       ErrorType = "permission_denied"
	ErrorTypeMissingBuildTools      ErrorType = "missing_build_tools"
	ErrorTypePackageManagerNotFound ErrorType = "package_manager_not_found"
	ErrorTypeArchitectureMismatch   ErrorType = "architecture_mismatch"
	ErrorTypeUnknown                ErrorType = "unknown"
)

type ErrorInfo struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Command string    `json:"command,omitempty"`
	Port    string    `json:"port,omitempty"`
	Package string    `json:"package,omitempty"`
}

func Parse(stderr string, exitCode int) *ErrorInfo {
	lowerStderr := strings.ToLower(stderr)

	if strings.Contains(lowerStderr, "command not found") || strings.Contains(lowerStderr, "executable file not found") {
		cmd := extractCommand(stderr)
		return &ErrorInfo{
			Type:    ErrorTypeMissingCommand,
			Message: "Command not found",
			Command: cmd,
		}
	}

	if strings.Contains(lowerStderr, "gcc") || strings.Contains(lowerStderr, "cc") || strings.Contains(lowerStderr, "compiler") {
		if strings.Contains(lowerStderr, "not found") || strings.Contains(lowerStderr, "no such file") {
			return &ErrorInfo{
				Type:    ErrorTypeMissingCompiler,
				Message: "Compiler not found",
			}
		}
	}

	if strings.Contains(lowerStderr, "cannot find -l") || strings.Contains(lowerStderr, "shared library") {
		pkg := extractLibrary(stderr)
		return &ErrorInfo{
			Type:    ErrorTypeMissingLibrary,
			Message: "Missing shared library",
			Package: pkg,
		}
	}

	if strings.Contains(lowerStderr, "address already in use") || strings.Contains(lowerStderr, "port is already in use") {
		port := extractPort(stderr)
		return &ErrorInfo{
			Type:    ErrorTypePortInUse,
			Message: "Port already in use",
			Port:    port,
		}
	}

	if strings.Contains(lowerStderr, "permission denied") {
		return &ErrorInfo{
			Type:    ErrorTypePermissionDenied,
			Message: "Permission denied",
		}
	}

	if strings.Contains(lowerStderr, "c compiler") || strings.Contains(lowerStderr, "make") {
		return &ErrorInfo{
			Type:    ErrorTypeMissingBuildTools,
			Message: "Missing build tools",
		}
	}

	if strings.Contains(lowerStderr, "package manager") && strings.Contains(lowerStderr, "not found") {
		return &ErrorInfo{
			Type:    ErrorTypePackageManagerNotFound,
			Message: "Package manager not found",
		}
	}

	return &ErrorInfo{
		Type:    ErrorTypeUnknown,
		Message: "Unknown error",
	}
}

func extractCommand(stderr string) string {
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "command not found") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				cmd := parts[len(parts)-1]
				if strings.HasPrefix(cmd, "'") || strings.HasPrefix(cmd, "\"") {
					cmd = cmd[1 : len(cmd)-1]
				}
				return cmd
			}
		}
	}
	return ""
}

func extractLibrary(stderr string) string {
	if idx := strings.Index(stderr, "-l"); idx >= 0 {
		rest := stderr[idx+2:]
		if i := strings.IndexAny(rest, " \n\t)"); i >= 0 {
			return rest[:i]
		}
		return rest
	}
	return ""
}

func extractPort(stderr string) string {
	parts := strings.Fields(stderr)
	for _, part := range parts {
		if strings.HasSuffix(part, ":") {
			num := part[:len(part)-1]
			if _, err := parseInt(num); err == nil && len(num) <= 5 {
				return num
			}
		}
	}
	return ""
}

func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result, nil
}
