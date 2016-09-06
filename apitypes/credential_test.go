package apitypes

import (
	"encoding/json"
	"testing"
)

func interfaceToCredentialValue(t *testing.T, i interface{}) (*CredentialValue, error) {
	b, err := json.Marshal(&i)
	if err != nil {
		t.Fatal("Encoding test data failed")
	}

	b, err = json.Marshal(string(b))
	if err != nil {
		t.Fatal("Encoding test data failed")
	}

	c := CredentialValue{}
	err = json.Unmarshal(b, &c)

	return &c, err
}

func TestCredentialValueUnmarshalJSON(t *testing.T) {
	t.Run("undefined", func(t *testing.T) {
		v := map[string]interface{}{
			"version": 1,
			"body": map[string]interface{}{
				"type":  "undefined",
				"value": "",
			},
		}

		c, err := interfaceToCredentialValue(t, v)
		if err != nil {
			t.Error("Unable to decode credential value: " + err.Error())
		}

		if !c.IsUnset() {
			t.Error("value was not considered unset")
		}
	})

	t.Run("string", func(t *testing.T) {
		v := map[string]interface{}{
			"version": 1,
			"body": map[string]interface{}{
				"type":  "string",
				"value": "test string",
			},
		}

		c, err := interfaceToCredentialValue(t, v)
		if err != nil {
			t.Error("Unable to decode credential value: " + err.Error())
		}

		if c.IsUnset() {
			t.Error("value is unset")
		}

		expected := "test string"
		if c.String() != expected {
			t.Errorf("wrong value! had: '%s' wanted: '%s'", c.String(), expected)
		}
	})

	t.Run("number (float)", func(t *testing.T) {
		v := map[string]interface{}{
			"version": 1,
			"body": map[string]interface{}{
				"type":  "number",
				"value": 108.5,
			},
		}

		c, err := interfaceToCredentialValue(t, v)
		if err != nil {
			t.Error("Unable to decode credential value: " + err.Error())
		}

		if c.IsUnset() {
			t.Error("value is unset")
		}

		expectedStr := "108.5"
		if c.String() != expectedStr {
			t.Errorf("wrong value! had: '%s' wanted: '%s'", c.String(), expectedStr)
		}

	})

	t.Run("number (int)", func(t *testing.T) {
		v := map[string]interface{}{
			"version": 1,
			"body": map[string]interface{}{
				"type":  "number",
				"value": 108,
			},
		}

		c, err := interfaceToCredentialValue(t, v)
		if err != nil {
			t.Error("Unable to decode credential value: " + err.Error())
		}

		if c.IsUnset() {
			t.Error("value is unset")
		}

		expectedStr := "108"
		if c.String() != expectedStr {
			t.Errorf("wrong value! had: '%s' wanted: '%s'", c.String(), expectedStr)
		}

	})
}
