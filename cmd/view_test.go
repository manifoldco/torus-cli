package cmd

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/pathexp"
)

type secretPair struct {
	key   string
	value string
}

func TestWriteEnvFormat(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	creds, path := viewCredentialsHelper(t)

	err := writeEnvFormat(w, creds, path)

	expected := `FOO=bar
BAZ="two words"
`
	w.Flush()
	got := string(buf.Bytes())

	if err != nil {
		t.Errorf("writeEnvFormat() expected no errors, got %s", err)
	}

	if expected != got {
		t.Errorf("writeEnvFormat() expected\n%q\ngot\n%q\n", expected, got)
	}
}

func TestWriteVerboseFormat(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	creds, path := viewCredentialsHelper(t)

	err := writeVerboseFormat(w, creds, path)

	expected := `Credential path: /o/p/e/s/*/i

FOO=bar          /o/p/e/s/*/i/foo
BAZ="two words"  /o/p/e/s/*/i/baz
`
	w.Flush()
	got := string(buf.Bytes())

	if err != nil {
		t.Errorf("writeVerboseFormat() expected no errors, got %s", err)
	}

	if expected != got {
		t.Errorf("writeVerboseFormat() expected\n%qgot\n%q", expected, got)
	}
}

func TestWriteJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	creds, path := viewCredentialsHelper(t)

	err := writeJSONFormat(w, creds, path)

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

func viewCredentialsHelper(t *testing.T) ([]apitypes.CredentialEnvelope, string) {
	var creds []apitypes.CredentialEnvelope

	pairs := []secretPair{
		{key: "foo", value: "bar"},
		{key: "baz", value: "two words"},
	}

	path := "/o/p/e/s/*/i"
	exp, err := pathexp.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range pairs {
		val := map[string]interface{}{
			"version": 2,
			"body": map[string]interface{}{
				"type":  "string",
				"value": s.value,
			},
		}

		cval, err := interfaceToCredentialValue(t, val)

		if err != nil {
			t.Fatal(err)
		}

		var cBody apitypes.Credential
		cBodyV2 := apitypes.CredentialV2{
			State: "set",
			BaseCredential: apitypes.BaseCredential{
				Name:    s.key,
				PathExp: exp,
				Value:   cval,
			},
		}
		cBody = &cBodyV2
		cred := apitypes.CredentialEnvelope{Body: &cBody}
		creds = append(creds, cred)
	}

	return creds, path
}
