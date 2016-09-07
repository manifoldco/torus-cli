package pathexp

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	type tc struct {
		path  string
		valid bool
	}
	testCases := []tc{}

	// happy case. a good path exp with no globbing or alternation.
	testCases = append(testCases,
		tc{path: "/org/project/env/service/user/instance", valid: true})

	// basic failure test case. incomplete pathexp
	parts := []string{"", "org", "project", "env", "service", "user"}
	for i := 0; i < len(parts); i++ {
		tc1 := tc{path: strings.Join(parts[:i], "/"), valid: false}
		tc2 := tc{path: strings.Join(parts[:i], "/") + "/", valid: false}
		testCases = append(testCases, tc1, tc2)
	}

	testCases = append(testCases,
		// too long pathexp
		tc{path: "/org/project/env/service/user/instance/soda", valid: false},
		// leading slash is missing
		tc{path: "org/project/env/service/user/instance", valid: false},
		// extra trailing slash
		tc{path: "/org/project/env/service/user/instance/", valid: false},
	)

	// Org and project can only be slugs
	testCases = append(testCases,
		tc{path: "/org-*/project/env/service/user/instance", valid: false},
		tc{path: "/*/project/env/service/user/instance", valid: false},
		tc{path: "/[ab|c]/project/env/service/user/instance", valid: false},

		tc{path: "/org/project-*/env/service/user/instance", valid: false},
		tc{path: "/org/*/env/service/user/instance", valid: false},
		tc{path: "/org/[ab|c]/env/service/user/instance", valid: false},
	)

	// Check globs and alternation for everything else. they should be valid
	multiple := []string{"*", "thing*", "a*", "[a|bc]", "[a|bc|d]", "[a*|c]"}
	prefix := "/org/project"
	parts = []string{"env", "service", "user", "instance"}
	for i := 0; i < len(parts); i++ {
		for _, glob := range multiple {
			globparts := []string{prefix}
			globparts = append(globparts, parts[:i]...)
			globparts = append(globparts, glob)
			globparts = append(globparts, parts[i+1:]...)
			path := strings.Join(globparts, "/")
			testCases = append(testCases, tc{path: path, valid: true})
		}
	}

	// Empty alternation
	testCases = append(testCases,
		tc{path: "/o/p/e/s/u/[]", valid: false},
	)

	// A too long part is invalid
	testCases = append(testCases, tc{
		path:  "/org/project/env/service/" + strings.Repeat("a", 80) + "/instance",
		valid: false,
	})

	// A full glob in alternation is invalid
	testCases = append(testCases, tc{
		path:  "/org/project/env/service/[*|foo]/instance",
		valid: false,
	})

	// Now run all the test cases
	for _, test := range testCases {
		t.Run(test.path, func(t *testing.T) {
			_, err := Parse(test.path)

			if (err == nil) != test.valid {
				t.Errorf("Test for %s failed. Expected validity = %t", test.path, test.valid)
			}
		})
	}

	// Check that what we put in is what we get out
	paths := []string{
		"/org/project/env/service/user/instance",
		"/org/project/e*/service/user/instance",
		"/org/project/env/*/user/instance",
		"/org/project/env/[abc|def]/user/instance",
		"/org/project/env/[abc|def|candy10]/user/instance",
		"/org/project/env/[abc|def|thing-*]/user/instance",
	}

	for _, path := range paths {
		t.Run("inverse fn check "+path, func(t *testing.T) {
			pe, err := Parse(path)
			if err != nil {
				t.Errorf("Parsing of %s failed", path)
			}

			out := pe.String()
			if path != out {
				t.Errorf("String() %s does not match input path %s", out, pe)
			}
		})
	}
}

func TestPathExpCompareSpecificity(t *testing.T) {
	type tc struct {
		a   string
		b   string
		res int
	}

	testCases := []tc{
		{a: "/o/p/e/s/u/i", b: "/o/p/e/s/u/i", res: 0},

		// check reciprocal arguments
		{a: "/o/p/e/s/u/i", b: "/o/p/*/s/u/i", res: 1},

		// check single glob
		{a: "/o/p/cool-env/s/u/i", b: "/o/p/cool-*/s/u/i", res: 1},

		// both globs, equal specificity
		{a: "/o/p/e/svc-*/u/i", b: "/o/p/e/sv*/u/i", res: 0},

		// alternation cases
		{a: "/o/p/e/svc/u/i", b: "/o/p/e/[boo|sv*]/u/i", res: 1},
		{a: "/o/p/e/sv*/u/i", b: "/o/p/e/[boo|sv*]/u/i", res: 1},
		{a: "/o/p/e/svc-*/u/i", b: "/o/p/e/[boo|sv*]/u/i", res: 1},
		{a: "/o/p/e/[svc-*|boo]/u/i", b: "/o/p/e/sv*/u/i", res: -1},
		{a: "/o/p/e/[svc-*|boo]/u/i", b: "/o/p/e/*/u/i", res: 1},
		{a: "/o/p/e/s/u/i", b: "/o/p/e/s/u*/i", res: 1},
	}

	for _, test := range testCases {
		t.Run(test.a+" "+test.b, func(t *testing.T) {
			a, err := Parse(test.a)
			if err != nil {
				t.Fatalf("Failed to parse %s", test.a)
			}

			b, err := Parse(test.b)
			if err != nil {
				t.Fatalf("Failed to parse %s", test.b)
			}

			res := a.CompareSpecificity(b)
			if res != test.res {
				t.Errorf("Expected %s CompareSpecificity %s = %d", test.a, test.b, test.res)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	type tc struct {
		a     string
		b     string
		equal bool
	}

	testCases := []tc{
		{a: "/o/p/e/s/u/i", b: "/o/p/e/s/u/i", equal: true},
		{a: "/o/p/*/s/u/i", b: "/o/p/*/s/u/i", equal: true},
		{a: "/o/p/e-*/s/u/i", b: "/o/p/e-*/s/u/i", equal: true},
		{a: "/o/p/e/[s|b]/u/i", b: "/o/p/e/[s|b]/u/i", equal: true},

		{a: "/o/p/e/s/u/i", b: "/o1/p/e/s/u/i", equal: false},
		{a: "/o/p/e/s/u/i", b: "/o/p1/e/s/u/i", equal: false},
		{a: "/o/p/e/s/u/i", b: "/o/p/*/s/u/i", equal: false},
		{a: "/o/p/e/s/u/i", b: "/o/p/e-*/s/u/i", equal: false},
		{a: "/o/p/e/s/u/i", b: "/o/p/e/[s|b]/u/i", equal: false},
		{a: "/o/p/e/s/u/i", b: "/o/p/e/[s|b-*]/u/i", equal: false},
		{a: "/o/p/e/[c|d|e]/u/i", b: "/o/p/e/[s|b-*]/u/i", equal: false},
		{a: "/o/p/e/[c|d|e]/u/i", b: "/o/p/e/[c|e|f]/u/i", equal: false},
	}

	for _, test := range testCases {
		t.Run(test.a+" "+test.b, func(t *testing.T) {
			a, err := Parse(test.a)
			if err != nil {
				t.Fatalf("Failed to parse %s", test.a)
			}

			b, err := Parse(test.b)
			if err != nil {
				t.Fatalf("Failed to parse %s", test.b)
			}

			res := a.Equal(b)
			rev := b.Equal(a)

			if res != rev {
				t.Error("(a == b) != (b == a)")
			}

			if res != test.equal {
				t.Errorf("Expected %s Equal %s = %t", test.a, test.b, test.equal)
			}
		})
	}
}

func TestWithInstance(t *testing.T) {
	a, err := Parse("/o/p/e/s/u/i")
	if err != nil {
		t.Fatalf("Failed to parse test item")
	}

	expected, err := Parse("/o/p/e/s/u/*")
	if err != nil {
		t.Fatalf("Failed to parse test item")
	}

	replaced, err := a.WithInstance("*")

	res := replaced.Equal(expected)
	if !res {
		t.Errorf("Expected %s Got %s", expected.String(), replaced.String())
	}
}
