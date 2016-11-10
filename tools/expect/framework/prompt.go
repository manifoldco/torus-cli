package framework

import (
	"bufio"
	"os"
	"time"

	"github.com/ThomasRooney/gexpect"

	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

const (
	inputChar = "↳ "
)

// Prompt is a set of fields that require input
type Prompt struct {
	Fields  []Field
	Timeout *time.Duration
}

// Execute runs the prompt
func (p Prompt) Execute(child *gexpect.ExpectSubprocess, output *Output) error {
	for _, f := range p.Fields {
		var timeout time.Duration
		if f.Timeout != nil {
			timeout = *f.Timeout
		} else {
			timeout = utils.GlobalTimeout
		}

		lbl := f.Label + ":"
		output.LogChild("? "+f.Label, 2)
		_, err := child.ExpectTimeoutRegexFind(lbl, timeout)
		if err != nil {
			return err
		}
		if f.RequestInput {
			output.Separator(3)
			readIn := makeReader(output)
			text := readIn("? Please enter value:\t", 3)
			output.LogChild(inputChar+text, 4)
			child.Send(text)
			output.Separator(3)
		} else if f.SendLine != "" {
			output.LogChild(inputChar+f.SendLine+"\\n", 3)
			child.Send(f.SendLine)
		} else {
			output.LogChild(inputChar+"\\n", 3)
		}
		child.Send("\n")
		err = expectValues(child, output, f.Expect, timeout, 1)
		if err != nil {
			return err
		}
	}
	return nil
}

// Field is a line item in a prompt
type Field struct {
	Label        string
	RequestInput bool
	SendLine     string
	Expect       []string
	Timeout      *time.Duration
}

func makeReader(output *Output) func(string, int) string {
	reader := bufio.NewReader(os.Stdin)

	return func(s string, tabLevel int) string {
		output.LogChild(s, tabLevel)
		txt, _ := reader.ReadString('\n')
		return txt
	}
}
