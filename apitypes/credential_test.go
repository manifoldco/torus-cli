package apitypes

import (
	"encoding/json"
	"strconv"
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
	tc := func(typ, val string, check func(*testing.T, *CredentialValue)) {
		t.Run(typ, func(t *testing.T) {
			v := map[string]interface{}{
				"version": 1,
				"body": map[string]interface{}{
					"type":  typ,
					"value": val,
				},
			}

			c, err := interfaceToCredentialValue(t, v)
			if err != nil {
				t.Error("Unable to decode credential value: " + err.Error())
			}

			check(t, c)
		})
	}

	tc("undefined", "", func(t *testing.T, c *CredentialValue) {
		if !c.IsUnset() {
			t.Error("value was not considered unset")
		}
	})

	tc("string", "test string", func(t *testing.T, c *CredentialValue) {
		if c.IsUnset() {
			t.Error("value is unset")
		}

		expected := "test string"
		if c.String() != expected {
			t.Errorf("wrong value! had: '%s' wanted: '%s'", c.String(), expected)
		}
	})

	tc("number", "108.5", func(t *testing.T, c *CredentialValue) {
		if c.IsUnset() {
			t.Error("value is unset")
		}

		expectedStr := "108.5"
		if c.String() != expectedStr {
			t.Errorf("wrong value! had: '%s' wanted: '%s'", c.String(), expectedStr)
		}

	})

	tc("number", "108", func(t *testing.T, c *CredentialValue) {
		if c.IsUnset() {
			t.Error("value is unset")
		}

		expectedStr := "108"
		if c.String() != expectedStr {
			t.Errorf("wrong value! had: '%s' wanted: '%s'", c.String(), expectedStr)
		}

	})

	tc("undecrypted", "", func(t *testing.T, c *CredentialValue) {
		if c.IsUnset() {
			t.Error("value is unset")
		}

		if !c.IsUndecrypted() {
			t.Error("value is undecrypted")
		}
	})

	t.Run("quoted json string", func(t *testing.T) {
		jsonString := strconv.Quote(`{"version":1,"body":{"type":"string","value":"12345678"}}`)
		c := CredentialValue{}
		err := json.Unmarshal([]byte(jsonString), &c)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if c.IsUnset() {
			t.Error("value is unset")
		}

		if c.String() != "12345678" {
			t.Errorf("wrong value! had '%s' wanted '%s'", c.String(), "12345678")
		}
	})
}
