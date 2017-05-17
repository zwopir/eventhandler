package runner

import (
	"bytes"
	"context"
	"github.com/prometheus/common/log"
	"io"
	"os/exec"
	"text/template"
	"time"
)

// PipeRunner represents a type that defines a command via an ExecFunc.
// Its Run method takes data as interface{} which are rendered an passed to the commands
// stdin io.Reader
type PipeRunner struct {
	Exec          ExecFunc
	StdinTemplate *template.Template
}

// NewPipeRunner creates a new PipeRunner
func NewPipeRunner(ctx context.Context, cmdString string, args []string, timeout time.Duration, tmpl *template.Template) *PipeRunner {
	execFunc := newExecFunc(ctx, cmdString, args, timeout)
	return &PipeRunner{
		Exec:          execFunc,
		StdinTemplate: tmpl,
	}
}

// Run connects the stdout io.Writer to the command, renders the provided data
// via PipeRunner.StdinTemplate and passes the result to the commands stdin
// The command stdout is written to the stdout io.Writer
//
// stdin -> PipeRunner.StdinTemplate -> ExecFunc -> stdout
func (pr *PipeRunner) Run(data interface{}, stdout io.Writer) error {
	var err error
	b := new(bytes.Buffer)
	err = pr.StdinTemplate.Execute(b, data)
	if err != nil {
		return err
	}
	log.Debugf("rendered stdin template to %s", b.String())
	err = pr.Exec(b, stdout)
	return err
}

// ExecFunc represents an adapter between a process, a stdin io.Reader and a
// stdout io.Writer
type ExecFunc func(stdinReader io.Reader, stdoutWriter io.Writer) error

// newExecFunc returns an ExecFunc with the Command set to os/exec.CommandContext
func newExecFunc(
	ctx context.Context,
	cmdString string,
	args []string,
	timeout time.Duration,
) ExecFunc {
	return func(r io.Reader, w io.Writer) error {
		ctx, done := context.WithTimeout(ctx, timeout)
		defer done()
		cmd := exec.CommandContext(ctx, cmdString, args...)
		cmd.Stdin = r
		cmd.Stdout = w
		return cmd.Run()
	}
}
