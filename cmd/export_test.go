package cmd

import (
	"bufio"
	"bytes"
	"testing"
)

func TestWriteJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	creds, _ := viewCredentialsHelper(t)

	err := writeJSONFormat(w, creds)

	expected := `{
  "baz": "two words",
  "foo": "bar"
}
`
	w.Flush()
	got := string(buf.Bytes())

	if err != nil {
		t.Errorf("writeJSONFormat() expected no errors, got %s", err)
	}

	if expected != got {
		t.Errorf("writeJSONFormat() expected\n%qgot\n%q", expected, got)
	}
}
