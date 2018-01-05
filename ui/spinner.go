package ui

import (
  "time"

  "github.com/briandowns/spinner"
)

type Spinner struct{
  spinner *spinner.Spinner
  text string
}

func NewSpinner(text string) *Spinner { return defUI.NewSpinner(text) }

func (u *UI) NewSpinner(text string) *Spinner {
  s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
  s.Suffix = " " + text
  return &Spinner {
    s,
    text,
  }
}

func (s *Spinner) Start() {
  s.spinner.Start()
}

func (s *Spinner) Stop() {
  s.spinner.Stop()
}

func (s *Spinner) Update(text string) {
  s.spinner.Suffix = " " + text
}
