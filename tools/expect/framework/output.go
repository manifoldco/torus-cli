package framework

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

const tab = "  "

// Output is a context-aware stdout printer
type Output struct {
	Context *ExpectContext
}

// Title outputs a title section
func (o *Output) Title(title string) {
	fmt.Println("")
	fmt.Println(title)
}

// Log outputs a string
func (o *Output) Log(text string) {
	fmt.Println(text)
}

// Separator outputs a 20 char line at a depth
func (o *Output) Separator(depth int) {
	o.LogChild(strings.Repeat("-", 20), depth)
}

// LogChild outputs a value at a tab depth
func (o *Output) LogChild(text string, depth int) {
	tabs := strings.Repeat(tab, depth)
	fmt.Println(tabs + text)
}

// CreateTabWriter wraps TabWriter.Init for ease of use
func (o *Output) CreateTabWriter(mw, tw, padding int, char byte, flags uint) TabWriter {
	t := TabWriter{}
	return t.Init(mw, tw, padding, char, flags)
}

// TabWriter is a wrapper to establish a TabWriter with options
type TabWriter struct {
	writer *tabwriter.Writer
}

// Init initializes the tabwriter.writer interface
func (t TabWriter) Init(mw, tw, padding int, char byte, flags uint) TabWriter {
	t.writer = tabwriter.NewWriter(os.Stdout, mw, tw, padding, char, flags)
	return t
}

// Write appends a line to a tabwriter
func (t TabWriter) Write(text string) {
	if t.writer == nil {
		panic("tabwriter not initialized")
	}
	fmt.Fprintln(t.writer, text)
}

// Flush finalizes the output of a tabwriter
func (t TabWriter) Flush() {
	if t.writer == nil {
		panic("tabwriter not initialized")
	}
	t.writer.Flush()
}
