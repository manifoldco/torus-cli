package cmd

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestImportSecretFile(t *testing.T) {
	t.Run("when argument file exists", func(t *testing.T) {
		tmpfile, err := ioutil.TempFile("", "secrets.txt")
		if err != nil {
			t.Fatal(err)
		}
		file := tmpfile.Name()

		content := `# this is an env file
FOO=bar
BAR=value # with comment
BAZ="multi word"
`

		err = ioutil.WriteFile(file, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}

		expect := []secretPair{
			{"FOO", "bar"},
			{"BAR", "value"},
			{"BAZ", "multi word"},
		}
		got, err := importSecretFile([]string{file})

		if err != nil {
			t.Errorf("importSecretFile(%v) expected no errors, got %q", file, err)
		}

		if !reflect.DeepEqual(expect, got) {
			t.Errorf("importSecretFile(%v) expected %q, got %q", file, expect, got)
		}
	})

	t.Run("when argument file doesn't exist", func(t *testing.T) {
		file := "torus-secret-missing-file.txt"

		expected := "torus-secret-missing-file.txt does not exist"
		_, got := importSecretFile([]string{file})

		if got.Error() != expected {
			t.Errorf("importSecretFile(%v) expected error %q, got %q", file, expected, got)
		}
	})

	t.Run("when multiple arguments are provided", func(t *testing.T) {
		files := []string{"a.env", "b.env"}

		expected := "Too many arguments were provided"
		_, got := importSecretFile(files)

		if got.Error() != expected {
			t.Errorf("importSecretFile(%v) expected error %q, got %q", files, expected, got)
		}
	})
}

func TestScanSecrets(t *testing.T) {
	type tc struct {
		content string
		pairs   []secretPair
		err     string
	}

	testCases := []tc{
		{
			content: `# comment line
				FOO=BAR
				bar=baz # comment at the end
				baz="multi value"

				/org/project/env/service/user/instance/foo=bar
			`,
			pairs: []secretPair{
				{"FOO", "BAR"},
				{"bar", "baz"},
				{"baz", "multi value"},
				{"/org/project/env/service/user/instance/foo", "bar"},
			},
		},
		{
			content: `FOO=BAR bar=baz # comment at the end`,
			pairs: []secretPair{
				{"FOO", "BAR"},
				{"bar", "baz"},
			},
		},
		{
			content: `foo`,
			err:     `Error parsing secret "foo"`,
		},
		{
			content: `foo = bar`,
			err:     `Error parsing secret "foo"`,
		},
		{
			content: `foo bar`,
			err:     `Error parsing secret "foo"`,
		},
		{
			content: `foo=`,
			err:     `Error parsing secret "foo="`,
		},
		{
			content: `=`,
			err:     `Error parsing secret "="`,
		},
		{
			content: `=bar`,
			err:     `Error parsing secret "=bar"`,
		},
	}

	for _, tc := range testCases {
		r := strings.NewReader(tc.content)
		pairs, err := scanSecrets(r)

		if err != nil && tc.err == "" {
			t.Errorf("scanSecrets(%v) expected no errors, got %q", tc.content, err)
		}

		if tc.err != "" && err.Error() != tc.err {
			t.Errorf("scanSecrets(%v) expected error %q, got %q", tc.content, tc.err, err)
		}

		if !reflect.DeepEqual(tc.pairs, pairs) {
			t.Errorf("scanSecrets(%v) expected %q, got %q", tc.content, tc.pairs, pairs)
		}
	}
}
