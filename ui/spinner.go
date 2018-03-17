package ui

import (
	"time"

	"github.com/briandowns/spinner"
)

// Spinner struct contains the spinner struct and display text
type Spinner struct {
	spinner *spinner.Spinner
	text    string
	enabled bool
}

// newSpinner creates a new Spinner struct
func newSpinner(text string, enabled bool) *Spinner {
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " " + text
	return &Spinner{
		s,
		text,
		enabled,
	}
}

// Start displays the Spinner and starts movement
func (s *Spinner) Start() {
	if !s.enabled {
		return
	}

	s.spinner.Start()
}

// Stop halts the spiner movement and removes it from display
func (s *Spinner) Stop() {
	if !s.enabled {
		return
	}

	s.spinner.Stop()
}

// Update changes the Spinner's suffix to 'text'
func (s *Spinner) Update(text string) {
	s.spinner.Suffix = " " + text
}
