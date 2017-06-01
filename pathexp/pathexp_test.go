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

	// Empty, single, and empty part alternation
	testCases = append(testCases,
		tc{path: "/o/p/e/s/u/[]", valid: false},
		tc{path: "/o/p/e/s/u/[i]", valid: false},
		tc{path: "/o/p/e/s/u/[|i]", valid: false},
		tc{path: "/o/p/e/s/u/[i|]", valid: false},
		tc{path: "/o/p/e/s/u/[|]", valid: false},
	)

	// partly formed alternations
	testCases = append(testCases,
		tc{path: "/o/p/e/s/u/[", valid: false},
		tc{path: "/o/p/e/s/u/[i", valid: false},
		tc{path: "/o/p/e/s/u/[i|", valid: false},
		tc{path: "/o/p/e/s/u/|]", valid: false},
		tc{path: "/o/p/e/s/u/|i]", valid: false},
		tc{path: "/o/p/e/s/u/b|i]", valid: false},
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

	// An empty segment is invalid
	testCases = append(testCases, tc{
		path:  "/org/project/env/service//instance",
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
		t.Fatal("Failed to parse test item")
	}

	expected, err := Parse("/o/p/e/s/u/*")
	if err != nil {
		t.Fatal("Failed to parse test item")
	}

	replaced, err := a.WithInstance("*")
	if err != nil {
		t.Fatal("Failed to replace instance")
	}

	res := replaced.Equal(expected)
	if !res {
		t.Errorf("Expected %s Got %s", expected.String(), replaced.String())
	}
}

func TestNormalize(t *testing.T) {
	path := "/org/project/env/[abc|def|candy10]/user/instance"
	pe, err := Parse(path)
	if err != nil {
		t.Errorf("Parsing of %s failed", path)
	}

	norm := "/org/project/env/[abc|candy10|def]/user/instance"
	out := pe.String()
	if out != norm {
		t.Errorf("String() %s does not match normalized form %s", out, norm)
	}
}

func TestDoubleGlob(t *testing.T) {
	path := "/org/project/**/instance"
	pe, err := Parse(path)
	if err != nil {
		t.Errorf("Parsing of %s failed", path)
	}

	norm := "/org/project/*/*/*/instance"
	out := pe.String()
	if out != norm {
		t.Errorf("String() %s does not match normalized form %s", out, norm)
	}

	path = "/org/project/**/identity/instance"
	pe, err = Parse(path)
	if err != nil {
		t.Errorf("Parsing of %s failed", path)
	}

	norm = "/org/project/*/*/identity/instance"
	out = pe.String()
	if out != norm {
		t.Errorf("String() %s does not match normalized form %s", out, norm)
	}
}

func TestSegmentsContains(t *testing.T) {
	path := "/org/project/[development|staging]/**/instance"
	pe, err := Parse(path)
	if err != nil {
		t.Errorf("Parsing of %s failed", path)
	}

	if !pe.Project.Contains("project") {
		t.Errorf("Literal contains failed to match match value")
	}

	if !pe.Envs.Contains("development") {
		t.Errorf("Alternation contains failed to match value")
	}

	if !pe.Envs.Contains("staging") {
		t.Errorf("Alternation contains failed to match value")
	}

	if !pe.Identities.Contains("development") {
		t.Errorf("FullGlob contains failed to match any value")
	}
}

func TestExpContains(t *testing.T) {
	testCases := []struct {
		l        string
		r        string
		contains bool
	}{
		{"/o/p/e/s/u/i", "/o/p/e/s/u/i", true},
		{"/o/p/*/s/u/i", "/o/p/e/s/u/i", true},
		{"/o/p/*/s/u/i", "/o/p/*/s/u/i", true},
		{"/o/p/e-*/s/u/i", "/o/p/e-*/s/u/i", true},
		{"/o/p/e/[s|b]/u/i", "/o/p/e/s/u/i", true},
		{"/o/p/e/[s|b]/u/i", "/o/p/e/b/u/i", true},
		{"/o/p/**", "/o/p/e/b/u/i", true},
		{"/o/p/**/i", "/o/p/e/b/u/i", true},

		{"/o/p/e1/s/u/i", "/o/p/e/s/u/i", false},
		{"/o/p/e/s/u/i", "/o/p/e1/s/u/i", false},
		{"/o/p/e/s/u/i", "/o/p/*/s/u/i", false},
		{"/o/p/e/s/u/i", "/o/p/e-*/s/u/i", false},
		{"/o/p/e/s/u/i", "/o/p/e/[s|b]/u/i", false},
		{"/o/p/e/[c|d|e]/u/i", "/o/p/e/s/u/i", false},
		{"/o/p/e/[c|d|e]/u/i", "/o/p/e/[c|d|e]/u/i", false},
	}

	for _, test := range testCases {
		t.Run(test.l+" "+test.r, func(t *testing.T) {
			l, err := Parse(test.l)
			if err != nil {
				t.Fatalf("Failed to parse %s", test.l)
			}

			r, err := Parse(test.r)
			if err != nil {
				t.Fatalf("Failed to parse %s", test.r)
			}

			if l.Contains(r) != test.contains {
				t.Errorf("Expected %s Contain %s = %t", test.l, test.r, test.contains)
			}
		})
	}
}
