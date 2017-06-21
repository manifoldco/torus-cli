package cmd

import (
	"testing"
)

func TestParseSetArgs(t *testing.T) {
	type tc struct {
		args  []string
		key   string
		value string
		err   string
	}

	testCases := []tc{
		{args: []string{}, err: "A secret name and value must be supplied"},
		{args: []string{"foo"}, err: "A secret name and value must be supplied"},
		{args: []string{"foo", "bar", "baz"}, err: "Too many arguments were provided"},
		{args: []string{"foo", "bar"}, key: "foo", value: "bar"},
		{args: []string{"foo=bar"}, key: "foo", value: "bar"},
		{args: []string{"foo=bar=="}, key: "foo", value: "bar=="},
		{args: []string{"="}, err: "A secret must have a name and value"},
		{args: []string{"key="}, err: "A secret must have a name and value"},
		{args: []string{"=sd"}, err: "A secret must have a name and value"},
		{
			args:  []string{"/org/project/env/service/user/instance/foo", "bar"},
			key:   "/org/project/env/service/user/instance/foo",
			value: "bar",
		},
		{
			args:  []string{"/org/project/env/service/user/instance/foo=bar"},
			key:   "/org/project/env/service/user/instance/foo",
			value: "bar",
		},
	}

	for _, tc := range testCases {
		key, value, err := parseSetArgs(tc.args)

		if err != nil && tc.err == "" {
			t.Errorf("parseSetArgs(%v) expected no errors, got %q", tc.args, err)
		}

		if tc.err != "" && err.Error() != tc.err {
			t.Errorf("parseSetArgs(%v) expected error %q, got %q", tc.args, tc.err, err)
		}

		if key != tc.key || value != tc.value {
			t.Errorf("parseSetArgs(%v) expected (%q,%q) pair, got (%q,%q)", tc.args,
				tc.key, tc.value, key, value)
		}
	}
}
