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

type PipeRunner struct {
	Exec          ExecFunc
	StdinTemplate *template.Template
}

func NewPipeRunner(ctx context.Context, cmdString string, args []string, timeout time.Duration, tmpl *template.Template) *PipeRunner {
	execFunc := newExecFunc(ctx, cmdString, args, timeout)
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
	log.Debugf("rendered stdin template to %s", data)
	err = pr.Exec(b, stdout)
	return err
}

type ExecFunc func(stdinReader io.Reader, stdoutWriter io.Writer) error

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
