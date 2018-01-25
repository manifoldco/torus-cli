package ui

import (
	"time"

	"github.com/briandowns/spinner"
)

// Spinner struct contains the spinner struct and display text
type Spinner struct {
	spinner *spinner.Spinner
	text    string
}

// NewSpinner creates a new Spinner struct
func NewSpinner(text string) *Spinner {
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " " + text
	return &Spinner{
		s,
		text,
	}
}

// Start displays the Spinner and starts movement
func (s *Spinner) Start() {
	s.spinner.Start()
}

// Stop halts the spiner movement and removes it from display
func (s *Spinner) Stop() {
	s.spinner.Stop()
}

// Update changes the Spinner's suffix to 'text'
func (s *Spinner) Update(text string) {
	s.spinner.Suffix = " " + text
}
