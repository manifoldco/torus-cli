package prompts

import (
	"os"
	"strconv"

	"github.com/manifoldco/promptui"
)

// SelectedAdd is returned when a user has requested to create the given value
var SelectedAdd = promptui.SelectedAdd

// selectWithAdd is a composite of promptui.Select and promptui.Prompt as
// promptui.SelectWithAdd does not support templating and several other
// features of promptui.Select.
type selectWithAdd struct {
	Label           string
	AddLabel        string
	Items           []string
	IsVimMode       bool
	Validate        promptui.ValidateFunc
	SelectTemplates *promptui.SelectTemplates
	PromptTemplates *promptui.PromptTemplates
}

func (sa *selectWithAdd) Run() (int, string, error) {
	if len(sa.Items) > 0 {
		newItems := append([]string{sa.AddLabel}, sa.Items...)

		s := promptui.Select{
			Label:     sa.Label,
			Items:     newItems,
			Templates: sa.SelectTemplates,
			IsVimMode: sa.IsVimMode,
		}

		idx, value, err := s.Run()
		if err != nil {
			return 0, "", err
		}

		if idx != 0 {
			return idx - 1, value, nil
		}

		os.Stdout.Write([]byte(upLine(1) + "\r" + clearLine))
	}

	p := promptui.Prompt{
		Label:     sa.AddLabel,
		Validate:  sa.Validate,
		Templates: sa.PromptTemplates,
		IsVimMode: sa.IsVimMode,
	}
	value, err := p.Run()
	return SelectedAdd, value, err
}

const esc = "\033["
const clearLine = esc + "2K"

func upLine(n uint) string {
	return movementCode(n, 'A')
}

func movementCode(n uint, code rune) string {
	return esc + strconv.FormatUint(uint64(n), 10) + string(code)
}
