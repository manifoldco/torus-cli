package prompts

import (
	"bytes"
	"text/template"

	"github.com/manifoldco/promptui"

	"github.com/manifoldco/torus-cli/errs"
)

const validationErrorTmpl = `  {{ . | faint }}`

func convertErr(err error) error {
	if err == promptui.ErrAbort || err == promptui.ErrInterrupt || err == promptui.ErrEOF {
		return errs.ErrAbort
	}

	return err
}

func direct(tpl string, data interface{}, value string) string {
	t, err := template.New("").Funcs(promptui.FuncMap).Parse(tpl)
	if err != nil {
		panic("could not parse template: " + err.Error())
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		panic("could not render template: " + err.Error())
	}

	buf.WriteString(value)
	return buf.String()
}
