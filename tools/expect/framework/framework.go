package framework

import (
	"time"

	"github.com/ThomasRooney/gexpect"

	"github.com/manifoldco/torus-cli/tools/expect/output"
	"github.com/manifoldco/torus-cli/tools/expect/utils"
)

// Command is a subcommand of an executable which has a list of expected outputs
// as well as data to satisfy any prompts that occur
type Command struct {
	Spawn    string
	Expect   []string
	Timeout  *time.Duration
	Prompt   *Prompt
	Callback func(data interface{})
}

// Execute spawns the command and runs through the expectations
func (c Command) Execute(executable string) error {
	output.Title(c.Spawn)
	command := executable + " " + c.Spawn
	child, err := gexpect.Spawn(command)
	defer child.Close()
	if err != nil {
		output.Log("Spawn errored")
		output.Log(err.Error())
		return err
	}

	// Fullfil the specified prompt
	if c.Prompt != nil {
		output.LogChild("Executing prompt", 1)
		err = c.Prompt.Execute(child)
		if err != nil {
			output.LogChild("Prompt failed", 2)
			return err
		}
	}

	var timeout time.Duration
	if c.Timeout != nil {
		timeout = *c.Timeout
	} else {
		timeout = utils.GlobalTimeout
	}

	// Final output from the command
	return expectValues(child, c.Expect, timeout, 1)
}

func expectValues(child *gexpect.ExpectSubprocess, values []string, timeout time.Duration, tabLevel int) error {
	if len(values) < 1 {
		return nil
	}

	output.LogChild("Expected output", tabLevel)
	for _, str := range values {
		_, err := child.ExpectTimeoutRegexFind(str, timeout)
		if err != nil {
			output.LogChild("✗ '"+str+"'", tabLevel+1)
			return err
		}
		output.LogChild("✔ '"+str+"'", tabLevel+1)
	}
	return nil
}
