package cmd

import (
	"encoding/json"
	"testing"

	"github.com/arigatomachine/cli/apitypes"
	"github.com/arigatomachine/cli/pathexp"
)

func interfaceToCredentialValue(t *testing.T, i interface{}) (*apitypes.CredentialValue, error) {
	b, err := json.Marshal(&i)
	if err != nil {
		t.Fatal("Encoding test data failed")
	}

	b, err = json.Marshal(string(b))
	if err != nil {
		t.Fatal("Encoding test data failed")
	}

	c := apitypes.CredentialValue{}
	err = json.Unmarshal(b, &c)

	return &c, err
}

func TestCredentialSet(t *testing.T) {

	uv := map[string]interface{}{
		"version": 1,
		"body": map[string]interface{}{
			"type":  "undefined",
			"value": "",
		},
	}

	unset, err := interfaceToCredentialValue(t, uv)
	if err != nil {
		t.Fatal("Unable to decode credential value: " + err.Error())
	}

	av := map[string]interface{}{
		"version": 1,
		"body": map[string]interface{}{
			"type":  "string",
			"value": "a",
		},
	}

	astring, err := interfaceToCredentialValue(t, av)
	if err != nil {
		t.Fatal("Unable to decode credential value: " + err.Error())
	}

	bv := map[string]interface{}{
		"version": 1,
		"body": map[string]interface{}{
			"type":  "string",
			"value": "b",
		},
	}

	bstring, err := interfaceToCredentialValue(t, bv)
	if err != nil {
		t.Fatal("Unable to decode credential value: " + err.Error())
	}

	t.Run("unset is ignored", func(t *testing.T) {
		cset := credentialSet{}

		path, _ := pathexp.Parse("/o/p/e/s/u/i")
		cred := apitypes.CredentialEnvelope{Body: &apitypes.Credential{
			Name:    "nothing",
			PathExp: path,
			Value:   unset,
		}}

		cset.Add(cred)

		if len(cset.ToSlice()) != 0 {
			t.Error("Unset credential was added to set")
		}
	})

	t.Run("most specific wins", func(t *testing.T) {

		path, _ := pathexp.Parse("/o/p/e/s/*/i")
		cred1 := apitypes.CredentialEnvelope{Body: &apitypes.Credential{
			Name:    "1",
			PathExp: path,
			Value:   astring,
		}}

		path, _ = pathexp.Parse("/o/p/e/s/u/i")
		cred2 := apitypes.CredentialEnvelope{Body: &apitypes.Credential{
			Name:    "1",
			PathExp: path,
			Value:   bstring,
		}}

		dotest := func(creds []apitypes.CredentialEnvelope) {
			cset := credentialSet{}
			for _, c := range creds {
				cset.Add(c)
			}

			slice := cset.ToSlice()
			if len(slice) != 1 {
				t.Errorf("Incorrect ToSlice length. wanted: %d got %d", 1, len(slice))
			}

			if slice[0].Body.Value != bstring {
				t.Error("Wrong value kept")
			}
		}

		dotest([]apitypes.CredentialEnvelope{cred1, cred2})
		dotest([]apitypes.CredentialEnvelope{cred2, cred1})
	})

	t.Run("output is sorted", func(t *testing.T) {

		makeCred := func(name string) apitypes.CredentialEnvelope {
			path, _ := pathexp.Parse("/o/p/e/s/*/i")
			return apitypes.CredentialEnvelope{Body: &apitypes.Credential{
				Name:    name,
				PathExp: path,
				Value:   astring,
			}}
		}

		cset := credentialSet{}

		cset.Add(makeCred("acred"))

		// these two should swap
		cset.Add(makeCred("ccred"))
		cset.Add(makeCred("bcred"))

		slice := cset.ToSlice()
		if len(slice) != 3 {
			t.Errorf("Incorrect ToSlice length. wanted: %d got %d", 1, len(slice))
		}

		if slice[0].Body.Name != "acred" || slice[1].Body.Name != "bcred" ||
			slice[2].Body.Name != "ccred" {

			t.Error("credentials not sorted")
		}
	})
}
