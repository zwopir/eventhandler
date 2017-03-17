package runner

import (
	"context"
	"io"
	"os/exec"
	"text/template"
	"bytes"
)

type PipeRunner struct {
	Exec          ExecFunc
	StdinTemplate *template.Template
}

func NewPipeRunner(ctx context.Context, cmdString string, args []string, tmpl *template.Template) *PipeRunner {
	execFunc := newExecFunc(ctx, cmdString, args)
	return &PipeRunner{
		Exec:          execFunc,
		StdinTemplate: tmpl,
	}
}

func (pr *PipeRunner) Run(data interface{}, stdout io.Writer) error {
	var err error
	b := bytes.NewBuffer([]byte(``))
	err = pr.StdinTemplate.Execute(b, &data)
	if err != nil {
		return err
	}
	err = pr.Exec(b, stdout)
	return err
}

type ExecFunc func(stdinReader io.Reader, stdoutWriter io.Writer) error

func newExecFunc(
	ctx context.Context,
	cmdString string,
	args []string,
) ExecFunc {
	return func(r io.Reader, w io.Writer) error {
		cmd := exec.CommandContext(ctx, cmdString, args...)
		cmd.Stdin = r
		cmd.Stdout = w
		cmd.Start()
		return cmd.Wait()
	}
}
