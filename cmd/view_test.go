package cmd

import (
	"testing"

	"github.com/manifoldco/torus-cli/apitypes"
	"github.com/manifoldco/torus-cli/pathexp"
)

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
