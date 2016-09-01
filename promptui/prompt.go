package promptui

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

var (
	bold  = styler(fgBold)
	faint = styler(fgFaint)
	red   = styler(fgBold, fgRed)

	initial = styler(fgBlue)("?")
	good    = styler(fgGreen)("✔")
	warn    = styler(fgYellow)("⚠")
	bad     = styler(fgRed)("✗")
)

// Prompt represents a single line text field input.
type Prompt struct {
	Label string // Label is the value displayed on the command line prompt

	Default string // Default is the initial value to populate in the prompt

	// Validate is optional. If set, this function is used to validate the input
	// after each character entry.
	Validate ValidateFunc

	// If mask is set, this value is displayed instead of the actual input
	// characters.
	Mask rune

	stdin  io.Reader
	stdout io.Writer
}

// Run runs the prompt, returning the validated input.
func (p *Prompt) Run() (string, error) {
	c := &readline.Config{}
	err := c.Init()
	if err != nil {
		return "", err
	}

	if p.stdin != nil {
		c.Stdin = p.stdin
	}

	if p.stdout != nil {
		c.Stdout = p.stdout
	}

	if p.Mask != 0 {
		c.EnableMask = true
		c.MaskRune = p.Mask
	}

	state := initial
	prompt := p.Label + ": "

	c.Prompt = bold(state) + " " + bold(prompt)
	c.HistoryLimit = -1
	c.UniqueEditLine = true

	firstListen := true
	wroteErr := false
	caughtup := true
	var out string

	if p.Default != "" {
		caughtup = false
		out = p.Default
		c.Stdin = io.MultiReader(bytes.NewBuffer([]byte(out)), os.Stdin)
	}

	rl, err := readline.NewEx(c)
	if err != nil {
		return "", err
	}

	validFn := func(x string) error {
		return nil
	}

	if p.Validate != nil {
		validFn = p.Validate
	}

	c.SetListener(func(line []rune, pos int, key rune) ([]rune, int, bool) {
		if firstListen {
			firstListen = false
			return nil, 0, false
		}

		if !caughtup && out != "" {
			if string(line) == out {
				caughtup = true
			}
			if wroteErr {
				return nil, 0, false
			}
		}

		err := validFn(string(line))
		if err != nil {
			if _, ok := err.(*ValidationError); ok {
				state = bad
			} else {
				rl.Close()
				return nil, 0, false
			}
		} else {
			state = good
		}

		rl.SetPrompt(bold(state) + " " + bold(prompt))
		rl.Refresh()
		wroteErr = false

		return nil, 0, false
	})

	for {
		out, err = rl.Readline()

		var msg string
		valid := true
		oerr := validFn(out)
		if oerr != nil {
			if verr, ok := oerr.(*ValidationError); ok {
				msg = verr.msg
				valid = false
				state = bad
			} else {
				return "", oerr
			}
		}

		if valid {
			state = good
			break
		}

		if err != nil {
			switch err {
			case readline.ErrInterrupt:
				err = errors.New("^C")
			case io.EOF:
				err = errors.New("^D")
			}

			break
		}

		caughtup = false

		c.Stdin = io.MultiReader(bytes.NewBuffer([]byte(out)), os.Stdin)
		rl, _ = readline.NewEx(c)

		firstListen = true
		wroteErr = true
		rl.SetPrompt("\n" + red(">> ") + msg + upLine(1) + "\r" + bold(state) + " " + bold(prompt))
		rl.Refresh()
	}

	if wroteErr {
		rl.Write([]byte(downLine(1) + clearLine() + upLine(1) + "\r"))
	}

	if err != nil {
		if err.Error() == "Interrupt" {
			err = errors.New("^C")
		}
		rl.Write([]byte("\n"))
		return "", err
	}

	echo := out
	if p.Mask != 0 {
		echo = strings.Repeat(string(p.Mask), len(echo))
	}

	rl.Write([]byte(state + " " + prompt + faint(echo) + "\n"))

	return out, err
}
