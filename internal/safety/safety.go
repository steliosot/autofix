package safety

var (
	allowlistCmds = map[string]bool{
		"npm":     true,
		"pip":     true,
		"pip3":    true,
		"python":  true,
		"python3": true,
		"node":    true,
		"docker":  true,
		"apt-get": true,
		"apt":     true,
		"dnf":     true,
		"yum":     true,
		"pacman":  true,
		"brew":    true,
		"curl":    true,
		"wget":    true,
		"git":     true,
		"make":    true,
		"gcc":     true,
		"clang":   true,
	}

	blockedCmds = map[string]bool{
		"rm -rf":   true,
		"rm -Rf":   true,
		"rm -r":    true,
		"rm -f /":  true,
		"userdel":  true,
		"usermod":  true,
		"mkfs":     true,
		"format":   true,
		"iptables": true,
		"ufw":      true,
		"firewall": true,
		"passwd":   true,
		"chpasswd": true,
		"shutdown": true,
		"reboot":   true,
	}
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(cmd string) error {
	cmdLower := toLower(cmd)

	for blocked := range blockedCmds {
		if contains(cmdLower, blocked) {
			return &ValidationError{Reason: "destructive command blocked: " + blocked}
		}
	}

	for allowed := range allowlistCmds {
		if hasPrefix(cmdLower, allowed+" ") || hasPrefix(cmdLower, allowed+"\t") {
			return nil
		}
	}

	if contains(cmdLower, "sudo") {
		return &ValidationError{Reason: "sudo commands require confirmation"}
	}

	return nil
}

func (v *Validator) IsLowRisk(cmd string) bool {
	cmdLower := toLower(cmd)

	for allowed := range allowlistCmds {
		if hasPrefix(cmdLower, allowed+" ") || hasPrefix(cmdLower, allowed+"\t") {
			return true
		}
	}

	return false
}

type ValidationError struct {
	Reason string
}

func (e *ValidationError) Error() string {
	return e.Reason
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + 32
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	n := len(substr)
	if n == 0 {
		return true
	}
	if n > len(s) {
		return false
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i] == substr[0] {
			match := true
			for j := 1; j < n; j++ {
				if s[i+j] != substr[j] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

func hasPrefix(s, prefix string) bool {
	n := len(prefix)
	if n > len(s) {
		return false
	}
	for i := 0; i < n; i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}
