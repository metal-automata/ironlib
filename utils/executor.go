package utils

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/metal-automata/ironlib/errs"
	"github.com/pkg/errors"
)

// Executor interface lets us implement dummy executors for tests
type Executor interface {
	Exec(context.Context) (*Result, error)
	ExecWithOutputWriter(ctx context.Context, w io.WriteCloser) (result *Result, err error)
	SetArgs(...string)
	SetEnv([]string)
	SetQuiet()
	SetVerbose()
	GetCmd() string
	DisableBinCheck()
	SetStdin(io.Reader)
	CmdPath() string
	CheckExecutable() error
	// for tests
	SetStdout([]byte)
	SetStderr([]byte)
	SetExitCode(int)
}

func NewExecutor(cmd string) Executor {
	return &Execute{Cmd: cmd, CheckBin: true}
}

// An execute instace
type Execute struct {
	Cmd      string
	Args     []string
	Env      []string
	Stdin    io.Reader
	CheckBin bool
	Quiet    bool
}

// The result of a command execution
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// GetCmd returns the command with args as a string
func (e *Execute) GetCmd() string {
	cmd := []string{e.Cmd}
	cmd = append(cmd, e.Args...)

	return strings.Join(cmd, " ")
}

// CmdPath returns the absolute path to the executable
// this means the caller should not have disabled CheckBin.
func (e *Execute) CmdPath() string {
	return e.Cmd
}

// SetArgs sets the command args
func (e *Execute) SetArgs(a ...string) {
	e.Args = a
}

// SetEnv sets the env variables
func (e *Execute) SetEnv(env []string) {
	e.Env = env
}

// SetQuiet lowers the verbosity
func (e *Execute) SetQuiet() {
	e.Quiet = true
}

// SetVerbose does whats it says
func (e *Execute) SetVerbose() {
	e.Quiet = false
}

// SetStdin sets the reader to the command stdin
func (e *Execute) SetStdin(r io.Reader) {
	e.Stdin = r
}

// DisableBinCheck disables validating the binary exists and is executable
func (e *Execute) DisableBinCheck() {
	e.CheckBin = false
}

// SetStdout doesn't do much, is around for tests
func (e *Execute) SetStdout(_ []byte) {
}

// SetStderr doesn't do much, is around for tests
func (e *Execute) SetStderr(_ []byte) {
}

// SetExitCode doesn't do much, is around for tests
func (e *Execute) SetExitCode(_ int) {
}

// Exec executes the command and returns the Result object
func (e *Execute) Exec(ctx context.Context) (result *Result, err error) {
	if e.CheckBin {
		err = e.CheckExecutable()
		if err != nil {
			return nil, err
		}
	}

	cmd := exec.CommandContext(ctx, e.Cmd, e.Args...) //nolint:gosec // This is a generic command so need to support this
	cmd.Env = append(cmd.Env, e.Env...)
	cmd.Stdin = e.Stdin

	var stdoutBuf, stderrBuf bytes.Buffer
	if !e.Quiet {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	} else {
		cmd.Stderr = &stderrBuf
		cmd.Stdout = &stdoutBuf
	}

	if err := cmd.Run(); err != nil {
		result = &Result{stdoutBuf.Bytes(), stderrBuf.Bytes(), cmd.ProcessState.ExitCode()}
		return result, newExecError(e.GetCmd(), result)
	}

	result = &Result{stdoutBuf.Bytes(), stderrBuf.Bytes(), cmd.ProcessState.ExitCode()}

	return result, nil
}

// CheckExecutable determines if the set Cmd value exists as a file and is an executable.
func (e *Execute) CheckExecutable() error {
	var path string

	if strings.Contains(e.Cmd, "/") {
		path = e.Cmd
	} else {
		var err error
		path, err = exec.LookPath(e.Cmd)
		if err != nil {
			return errors.Wrap(errs.ErrBinLookupPath, err.Error())
		}

		e.Cmd = path
	}

	fileInfo, err := os.Lstat(path)
	if err != nil {
		return errors.Wrap(errs.ErrBinLstat, err.Error())
	}

	// bit mask 0111 indicates atleast one of owner, group, others has an executable bit set
	if fileInfo.Mode()&0o111 == 0 {
		return errs.ErrBinNotExecutable
	}

	return nil
}

// ExecWithStoutPipe executes the command while writing the output - stdout and stderr to the given pipe
//
// Note: The returned result.Stdout, result.Stderr fields are not populated
func (e *Execute) ExecWithOutputWriter(ctx context.Context, w io.WriteCloser) (result *Result, err error) {
	defer w.Close()
	if e.CheckBin {
		err = e.CheckExecutable()
		if err != nil {
			return nil, err
		}
	}

	cmd := exec.CommandContext(ctx, e.Cmd, e.Args...) //nolint:gosec // yes, this is expected - G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
	cmd.Env = append(cmd.Env, e.Env...)
	cmd.Stdin = e.Stdin

	destStdout := []io.Writer{w}
	destStderr := []io.Writer{w}

	// include stdout, stderr on verbose
	if !e.Quiet {
		destStdout = append(destStdout, os.Stdout)
		destStderr = append(destStderr, os.Stderr)
	}

	cmd.Stdout = io.MultiWriter(destStdout...)
	cmd.Stderr = io.MultiWriter(destStderr...)

	if err := cmd.Run(); err != nil {
		result = &Result{ExitCode: cmd.ProcessState.ExitCode()}
		return result, newExecError(e.GetCmd(), result)
	}

	result = &Result{ExitCode: cmd.ProcessState.ExitCode()}

	return result, nil
}
