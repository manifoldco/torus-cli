package cmd

import "testing"

func TestParseArgs(t *testing.T) {
	type tc struct {
		args []string
		key  string
		val  string
		err  string
	}

	testCases := []tc{
		{args: []string{}, err: "A secret name and value must be supplied."},
		{args: []string{"foo"}, err: "A secret name and value must be supplied."},
		{args: []string{"foo", "bar", "baz"}, err: "Too many arguments were provided."},
		{args: []string{"foo", "bar"}, key: "foo", val: "bar"},
		{args: []string{"foo=bar"}, key: "foo", val: "bar"},
		{args: []string{"foo=bar=="}, key: "foo", val: "bar=="},
		{args: []string{"="}, err: "A secret must have a name and value."},
		{args: []string{"key="}, key: "key", err: "A secret must have a name and value."},
		{args: []string{"=sd"}, val: "sd", err: "A secret must have a name and value."},
		{
			args: []string{"/org/project/env/service/user/instance/foo", "bar"},
			key:  "/org/project/env/service/user/instance/foo",
			val:  "bar",
		},
		{
			args: []string{"/org/project/env/service/user/instance/foo=bar"},
			key:  "/org/project/env/service/user/instance/foo",
			val:  "bar",
		},
	}

	for _, tc := range testCases {
		key, val, err := parseSetArgs(tc.args)

		if err != "" && tc.err == "" {
			t.Errorf("parseArgs(%v) expected no errors, got %q", tc.args, err)
		}

		if tc.err != "" && err != tc.err {
			t.Errorf("parseArgs(%v) expected error %q, got %q", tc.args, tc.err, err)
		}

		if key != tc.key || val != tc.val {
			t.Errorf("parseArgs(%v) expected (%q,%q) pair, got (%q,%q)", tc.args,
				tc.key, tc.val, key, val)
		}
	}
}
