package executor

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"
)

type Result struct {
	Command  string   `json:"command"`
	ExitCode int      `json:"exit_code"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	Success  bool     `json:"success"`
	Lines    []string `json:"lines"`
}

var Runner = func(cmd *exec.Cmd) (*Result, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Command: strings.Join(cmd.Args, " "),
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Lines:   strings.Split(stdout.String(), "\n"),
	}

	if err != nil {
		result.ExitCode = getExitCode(err)
		result.Success = false
	} else {
		result.ExitCode = 0
		result.Success = true
	}

	return result, nil
}

func getExitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 1
}
