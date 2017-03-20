package runner

import (
	"testing"
	"text/template"
	"context"
	"io"
	"bytes"
	"bufio"
)


func createTestTemplate() *template.Template {
	const templateText= `{{ index . "key"}}`
	tmpl, _ := template.New("testTemplate").Parse(templateText)
	return tmpl
}

var runnerTestTable = []struct {
	data map[string]string
	cmdString string
	args []string
	template *template.Template
}{
	{
		map[string]string{"key": "value"},
		"cat",
		[]string{"-"},
		createTestTemplate(),
	},
}

func TestNewPipeRunner(t *testing.T) {
	for _, tt := range runnerTestTable {
		_ = NewPipeRunner(
			context.Background(),
			tt.cmdString,
			tt.args,
			tt.template,
		)
	}
}

func createMockExecFunc() ExecFunc {
	return func(r io.Reader, w io.Writer) error {
		b := bufio.NewReader(r)
		b.WriteTo(w)
		return nil
	}
}


var runTestTable = []struct{
	data map[string]string
	template *template.Template
	expectedOutput []byte
	execFunc ExecFunc
}{
	{
		map[string]string{"key": "value"},
		createTestTemplate(),
		[]byte(`value`),
		createMockExecFunc(),
	},
}

func TestPipeRunner_Run(t *testing.T) {
	for _, tt := range runTestTable {
		pr := &PipeRunner{
			Exec: tt.execFunc,
			StdinTemplate: tt.template,
		}
		out := bytes.NewBuffer([]byte(``))
		err := pr.Run(tt.data, out)
		if err != nil {
			t.Errorf("running mock exec func returned an error: %s", err)
		}
		if string(out.Bytes()) != string(tt.expectedOutput) {
			t.Errorf("expected %s as result from mock exec, got %s",
				tt.expectedOutput, out.Bytes(),
			)
		}
	}
}


func TestPipeRunner_Run2(t *testing.T) {
	for _, tt := range runnerTestTable {
		pr := NewPipeRunner(
			context.Background(),
			tt.cmdString,
			tt.args,
			tt.template,
		)
		b := bytes.NewBuffer([]byte(``))
		err := pr.Run(tt.data, b)
		if err != nil {
			t.Errorf("running %s returned an error: %s",
				tt.cmdString, err,
			)
		}

		if string(b.Bytes()) != tt.data["key"] {
			t.Errorf("expected %s from running %s, got %s",
				tt.data["key"], string(b.Bytes()),
			)
		}
	}
}

