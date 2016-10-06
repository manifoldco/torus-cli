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

	// Version one unset credential
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

	// Version two unset credential
	uv2 := map[string]interface{}{
		"version": 2,
		"state":   "unset",
		"body": map[string]interface{}{
			"type":  "undefined",
			"value": "",
		},
	}

	unsetv2, err := interfaceToCredentialValue(t, uv2)
	if err != nil {
		t.Fatal("Unable to decode credential value: " + err.Error())
	}

	av := map[string]interface{}{
		"version": 2,
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
		"version": 2,
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

		// Version two unset credential
		path, _ := pathexp.Parse("/o/p/e/s/*/i")
		var cBody apitypes.Credential
		cBodyV2 := apitypes.CredentialV2{
			State: "unset",
			BaseCredential: apitypes.BaseCredential{
				Name:    "nothing",
				PathExp: path,
				Value:   unsetv2,
			},
		}
		cBody = &cBodyV2
		cred := apitypes.CredentialEnvelope{Body: &cBody}
		cset.Add(cred)

		// Version one unset credential
		path, _ = pathexp.Parse("/o/p/e/s/u/i")
		cBodyV1 := apitypes.BaseCredential{
			Name:    "nothing",
			PathExp: path,
			Value:   unset,
		}
		cBody = &cBodyV1
		cred = apitypes.CredentialEnvelope{Body: &cBody}
		cset.Add(cred)

		if len(cset.ToSlice()) != 0 {
			t.Error("Unset credential was added to set")
		}
	})

	t.Run("most specific wins", func(t *testing.T) {

		path, _ := pathexp.Parse("/o/p/e/s/*/i")
		var cBodyOne apitypes.Credential
		cBodyOneV2 := apitypes.CredentialV2{
			State: "set",
			BaseCredential: apitypes.BaseCredential{
				Name:    "1",
				PathExp: path,
				Value:   astring,
			},
		}
		cBodyOne = &cBodyOneV2
		credOne := apitypes.CredentialEnvelope{Body: &cBodyOne}

		path, _ = pathexp.Parse("/o/p/e/s/u/i")
		var cBodyTwo apitypes.Credential
		cBodyTwoV2 := apitypes.CredentialV2{
			State: "set",
			BaseCredential: apitypes.BaseCredential{
				Name:    "1",
				PathExp: path,
				Value:   bstring,
			},
		}
		cBodyTwo = &cBodyTwoV2
		credTwo := apitypes.CredentialEnvelope{Body: &cBodyTwo}

		dotest := func(creds []apitypes.CredentialEnvelope) {
			cset := credentialSet{}
			for _, c := range creds {
				cset.Add(c)
			}

			slice := cset.ToSlice()
			if len(slice) != 1 {
				t.Errorf("Incorrect ToSlice length. wanted: %d got %d", 1, len(slice))
			}

			if (*slice[0].Body).GetValue() != bstring {
				t.Error("Wrong value kept")
			}
		}

		dotest([]apitypes.CredentialEnvelope{credOne, credTwo})
		dotest([]apitypes.CredentialEnvelope{credTwo, credOne})
	})

	t.Run("output is sorted", func(t *testing.T) {

		makeCred := func(name string) apitypes.CredentialEnvelope {
			path, _ := pathexp.Parse("/o/p/e/s/*/i")
			var cBody apitypes.Credential
			cBodyV2 := apitypes.CredentialV2{
				State: "set",
				BaseCredential: apitypes.BaseCredential{
					Name:    name,
					PathExp: path,
					Value:   astring,
				},
			}
			cBody = &cBodyV2
			return apitypes.CredentialEnvelope{Body: &cBody}
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

		one := *slice[0].Body
		two := *slice[1].Body
		three := *slice[2].Body
		if one.GetName() != "acred" || two.GetName() != "bcred" || three.GetName() != "ccred" {
			t.Error("credentials not sorted")
		}
	})
}
