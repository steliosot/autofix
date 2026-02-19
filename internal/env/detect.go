package env

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func Detect() *Environment {
	env := &Environment{
		OS:           detectOS(),
		OSVersion:    detectOSVersion(),
		Architecture: detectArchitecture(),
	}
	env.PackageManager = detectPackageManager(env.OS)
	env.HasSudo = detectSudo()
	env.InContainer = detectContainer()
	env.Runtimes = detectRuntimes()
	return env
}

func detectOS() OS {
	if runtime.GOOS == "darwin" {
		return OSMacOS
	}

	for _, path := range []string{"/etc/os-release", "/etc/lsb-release"} {
		content, err := os.ReadFile(path)
		if err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "Ubuntu") {
				return OSUbuntu
			}
			if strings.Contains(contentStr, "Debian") {
				return OSDebian
			}
			if strings.Contains(contentStr, "Fedora") {
				return OSFedora
			}
		}
	}

	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return OSArch
	}

	if _, err := os.Stat("/etc/fedora-release"); err == nil {
		return OSFedora
	}

	return OSUnknown
}

func detectOSVersion() string {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("sw_vers", "-productVersion")
		if output, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}

	content, err := os.ReadFile("/etc/os-release")
	if err == nil {
		contentStr := string(content)
		for _, line := range strings.Split(contentStr, "\n") {
			if strings.HasPrefix(line, "VERSION=") {
				return strings.Trim(strings.TrimPrefix(line, "VERSION="), `"`)
			}
		}
	}

	return "unknown"
}

func detectArchitecture() Architecture {
	arch := runtime.GOARCH
	if arch == "amd64" {
		return ArchAMD64
	}
	if arch == "arm64" {
		return ArchARM64
	}
	return ArchUnknown
}

func detectPackageManager(os OS) PackageManager {
	switch os {
	case OSMacOS:
		if _, err := exec.LookPath("brew"); err == nil {
			return PMBrew
		}
	case OSUbuntu, OSDebian:
		if _, err := exec.LookPath("apt"); err == nil {
			return PMApt
		}
	case OSFedora:
		if _, err := exec.LookPath("dnf"); err == nil {
			return PMDnf
		}
		if _, err := exec.LookPath("yum"); err == nil {
			return PMYum
		}
	case OSArch:
		if _, err := exec.LookPath("pacman"); err == nil {
			return PMPacman
		}
	}
	return PMNone
}

func detectSudo() bool {
	_, err := exec.LookPath("sudo")
	return err == nil
}

func detectContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	content, err := os.ReadFile("/proc/1/cgroup")
	if err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "docker") || strings.Contains(contentStr, "lxc") {
			return true
		}
	}
	return false
}

func detectRuntimes() []Runtime {
	runtimes := []Runtime{}

	if version, err := detectNodeVersion(); err == nil {
		runtimes = append(runtimes, Runtime{Name: "node", Version: version})
	}

	if version, err := detectPythonVersion(); err == nil {
		runtimes = append(runtimes, Runtime{Name: "python", Version: version})
	}

	if version, err := detectDockerVersion(); err == nil {
		runtimes = append(runtimes, Runtime{Name: "docker", Version: version})
	}

	return runtimes
}

func detectNodeVersion() (string, error) {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func detectPythonVersion() (string, error) {
	cmd := exec.Command("python3", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func detectDockerVersion() (string, error) {
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
