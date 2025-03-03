package utils

import (
	"errors"
	"fmt"
)

var (
	ErrNoCommandOutput          = errors.New("command returned no output")
	ErrVersionStrExpectedSemver = errors.New("expected version string to follow semver format")
	ErrFakeExecutorInvalidArgs  = errors.New("invalid number of args passed to fake executor")
	ErrRepositoryBaseURL        = errors.New("repository base URL undefined, ensure UpdateOptions.BaseURL OR UPDATE_BASE_URL env var is set")
	ErrRebootRequired           = errors.New("reboot required")
)

// ExecError is returned when the command exits with an error or a non zero exit status
type ExecError struct {
	Cmd      string
	Stderr   string
	Stdout   string
	ExitCode int
}

// Error implements the error interface
func (u *ExecError) Error() string {
	errMsg := fmt.Sprintf("'%s' exited with exit code: %d", u.Cmd, u.ExitCode)

	if u.Stderr != "" {
		errMsg += fmt.Sprintf(", stderr: %s", u.Stderr)
	}

	if u.Stdout != "" {
		errMsg += fmt.Sprintf(", stdout: %s", u.Stdout)
	}

	return errMsg
}

func newExecError(cmd string, r *Result) *ExecError {
	return &ExecError{
		Cmd:      cmd,
		Stderr:   string(r.Stderr),
		Stdout:   string(r.Stdout),
		ExitCode: r.ExitCode,
	}
}
