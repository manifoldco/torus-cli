/*
Package pathexp provides a representation of path expressions; locations of
secrets within the org/project/environment/service/identity/instance hierarchy
supporting globs and alternation.

Grammar:

	<pathexp>     ::= "/" <org> "/" <project> "/" <environment> "/" <service> "/" <identity> "/" <instance>
	<org>         ::= <literal>
	<project>     ::= <literal>
	<environment> ::= <multiple>
	<service>     ::= <multiple>
	<identity>    ::= <multiple>
	<instance>    ::= <multiple>

	<multiple>         ::= <alternation> | <glob-or-literal> | <full-glob>
	<alternation>      ::= "[" <alternation-body> "]"
	<alternation-body> ::= <glob-or-literal> | <glob-or-literal> "|" <alternation-body>
	<glob-or-literal>  ::= <glob> | <literal>
	<glob>             ::= <literal> "*"
	<literal>          ::= [a-z0-9][a-z0-9\-\_]{0,63}
	<fullglob>         ::= "*"
*/
package pathexp

import (
	"errors"
	"regexp"
	"sort"
	"strings"

	"github.com/manifoldco/torus-cli/errs"
)

const slugstr = `[a-z\d][-_a-z\d]{0,63}`

var (
	slug           = regexp.MustCompile(`^` + slugstr + `$`)
	globRe         = regexp.MustCompile(`^(` + slugstr + `)(\*?)$`)
	fullglobOrGlob = regexp.MustCompile(`(^\*$)|(?:^(` + slugstr + `)(\*?)$)`)
)

const (
	orgIdx = iota
	projectIdx
	envIdx
	serviceIdx
	identityIdx
	instanceIdx
)

// PathExp is a path expression
type PathExp struct {
	org        literal
	project    literal
	envs       segment
	services   segment
	identities segment
	instances  segment
}

type segment interface {
	String() string
}

type literal string
type glob string
type fullglob struct{}
type alternation []segment

func (l literal) String() string  { return string(l) }
func (g glob) String() string     { return string(g) + "*" }
func (f fullglob) String() string { return "*" }
func (a alternation) String() string {
	strs := []string{}
	for _, s := range a {
		strs = append(strs, s.String())
	}

	// XXX: pathexps must be normalized in sorted order for the server to accept
	// them. We should revist this, so what the user puts in is what they get
	// out.
	sort.Strings(strs)

	return "[" + strings.Join(strs, "|") + "]"
}

// compareSegmentType ranks the segments by their type specificity
func compareSegmentType(a, b segment) int {
	segs := []segment{a, b}
	ranks := make([]int, 2)
	for i, seg := range segs {
		switch seg.(type) {
		case literal:
			ranks[i] = 3
		case glob:
			ranks[i] = 2
		case alternation:
			ranks[i] = 1
		case fullglob:
			ranks[i] = 0
		default:
			panic("Bad type for segment!")
		}
	}

	switch {
	case ranks[0] < ranks[1]:
		return -1
	case ranks[0] > ranks[1]:
		return 1
	default:
		return 0
	}
}

func segmentsEqual(a, b segment) bool {
	switch at := a.(type) {
	case literal:
		if bl, ok := b.(literal); ok {
			return at == bl
		}
		return false
	case glob:
		if bg, ok := b.(glob); ok {
			return at == bg
		}
		return false
	case alternation:
		if ba, ok := b.(alternation); ok {
			if len(at) != len(ba) {
				return false
			}

		LoopA:
			for _, av := range at {
				for _, bv := range ba {
					if segmentsEqual(av, bv) {
						continue LoopA
					}
				}
				return false
			}

			return true
		}
		return false

	case fullglob:
		_, ok := b.(fullglob)
		return ok
	default:
		panic("Bad type for segment!")
	}
}

// NewPartial creates a new path expression from the given path segments
// It returns an error if any of the values fail to validate
func NewPartial(org, project string, envs, services, identities, instances []string) (*PathExp, *int, error) {
	orgValue := org
	if org == "" {
		orgValue = "*"
	}
	projValue := project
	if project == "" {
		projValue = "*"
	}

	pe := PathExp{
		org:     literal(orgValue),
		project: literal(projValue),
	}
	segmentCount := 0

	if org != "" && org != "*" && !slug.MatchString(org) {
		return nil, nil, errors.New("Invalid org")
	} else if org != "" {
		segmentCount++
	}

	if project != "" && project != "*" && !slug.MatchString(project) {
		return nil, nil, errors.New("Invalid project: " + project)
	} else if project != "" {
		segmentCount++
	}

	var err error

	pe.envs, err = parseMultiple("environment", envs)
	if err != nil {
		if !strings.Contains(err.Error(), "Empty") {
			return nil, nil, err
		}
		pe.envs = fullglob{}
	} else if pe.envs.String() != "" {
		segmentCount++
	}

	pe.services, err = parseMultiple("service", services)
	if err != nil {
		if !strings.Contains(err.Error(), "Empty") {
			return nil, nil, err
		}
		pe.services = fullglob{}
	} else if pe.services.String() != "" {
		segmentCount++
	}
	var originalEnv segment
	if pe.services.String() != "*" && pe.envs.String() != "*" {
		originalEnv = pe.envs
		pe.envs = fullglob{}
	}

	pe.identities, err = parseMultiple("identity", identities)
	if err != nil {
		if !strings.Contains(err.Error(), "Empty") {
			return nil, nil, err
		}
		pe.identities = fullglob{}
	} else if pe.identities.String() != "" {
		segmentCount++
	}

	pe.instances, err = parseMultiple("instance", instances)
	if err != nil {
		if !strings.Contains(err.Error(), "Empty") {
			return nil, nil, err
		}
		pe.instances = fullglob{}
	} else if pe.instances.String() != "" {
		segmentCount++
	}
	if segmentCount > 3 && originalEnv != nil {
		pe.envs = originalEnv
	}

	return &pe, &segmentCount, nil
}

// New creates a new path expression from the given path segments
// It returns an error if any of the values fail to validate
func New(org, project string, envs, services, identities, instances []string) (*PathExp, error) {
	pe := PathExp{
		org:     literal(org),
		project: literal(project),
	}

	if !slug.MatchString(org) {
		return nil, errors.New("invalid org name")
	}

	if !slug.MatchString(project) {
		return nil, errors.New("invalid project name")
	}

	var err error

	pe.envs, err = parseMultiple("environment", envs)
	if err != nil {
		return nil, err
	}

	pe.services, err = parseMultiple("service", services)
	if err != nil {
		return nil, err
	}

	pe.identities, err = parseMultiple("identity", identities)
	if err != nil {
		return nil, err
	}

	pe.instances, err = parseMultiple("instance", instances)
	if err != nil {
		return nil, err
	}

	return &pe, nil
}

// Parse parses a string into a path expression. It returns an error if the
// string is not a valid path expression.
func Parse(raw string) (*PathExp, error) {
	parts := strings.Split(raw, "/")

	if len(parts) != 7 {
		return nil, errors.New("wrong number of path segments")
	}

	if parts[0] != "" {
		return nil, errors.New("path expressions must start with '/'")
	}
	// remove leading empty section
	parts = parts[1:]

	splitParts := make([][]string, 6)
	splitNames := []string{"", "", "environment", "service", "identity", "instance"}
	var err error
	for i := 2; i < len(splitParts); i++ {
		splitParts[i], err = Split(splitNames[i], parts[i])
		if err != nil {
			return nil, err
		}
	}

	return New(parts[orgIdx], parts[projectIdx],
		splitParts[envIdx],
		splitParts[serviceIdx],
		splitParts[identityIdx],
		splitParts[instanceIdx],
	)
}

// WithInstance clones a PathExp, replacing its instance with the parsed value
// from the argument.
//
// XXX: this isn't really great. it would be nice to support all path types.
func (pe *PathExp) WithInstance(instance string) (*PathExp, error) {
	parts, err := Split("instance", instance)
	if err != nil {
		return nil, err
	}

	segment, err := parseMultiple("instance", parts)
	if err != nil {
		return nil, err
	}

	return &PathExp{
		org:        pe.org,
		project:    pe.project,
		envs:       pe.envs,
		services:   pe.services,
		identities: pe.identities,
		instances:  segment,
	}, nil
}

// Org returns the org set for this pathexp
func (pe *PathExp) Org() string {
	return string(pe.org)
}

// Project returns the project set for this pathexp
func (pe *PathExp) Project() string {
	return string(pe.project)
}

// Envs returns the envs set for this pathexp
func (pe *PathExp) Envs() string {
	return pe.envs.String()
}

// Services returns the services set for this pathexp
func (pe *PathExp) Services() string {
	return pe.services.String()
}

// Identities returns the identities set for this pathexp
func (pe *PathExp) Identities() string {
	return pe.identities.String()
}

// Instances returns the instances set for this pathexp
func (pe *PathExp) Instances() string {
	return pe.instances.String()
}

// String returns the unparsed string representation of the path expression
func (pe *PathExp) String() string {
	return strings.Join([]string{"", string(pe.org), string(pe.project),
		pe.envs.String(),
		pe.services.String(),
		pe.identities.String(),
		pe.instances.String(),
	}, "/")
}

// Split separates alternation
func Split(name, segment string) ([]string, error) {
	parts := []string{segment}

	if len(segment) == 0 {
		return parts, nil // let elsewhere handle the empty single segment
	}

	if segment[0] == '[' && segment[len(segment)-1] == ']' {
		parts = strings.Split(segment[1:len(segment)-1], "|")
		// zero length is checked in parseMultiple
		if len(parts) == 1 {
			return nil, errors.New("Single item in segment alternation for " + name + ".")
		}
	}

	return parts, nil
}

func parseMultiple(name string, parts []string) (segment, error) {
	switch len(parts) {
	case 0:
		return nil, errors.New("Empty segment alternation for " + name + ".")
	case 1:
		matches := fullglobOrGlob.FindAllStringSubmatch(parts[0], -1)
		if len(matches) != 1 {
			return nil, errors.New("Invalid " + name + ".")
		}

		match := matches[0]
		switch {
		case match[1] != "": // fullglob
			return fullglob{}, nil
		case match[3] != "": // glob
			return glob(match[2]), nil
		default: // literal
			return literal(match[2]), nil
		}
	default:
		var res alternation
		for _, part := range parts {
			matches := globRe.FindAllStringSubmatch(part, -1)
			if len(matches) != 1 {
				return nil, errors.New("Invalid " + name + ".")
			}

			match := matches[0]
			switch {
			case match[2] != "": // glob
				res = append(res, glob(match[1]))
			default: // literal
				res = append(res, literal(match[1]))
			}
		}

		return res, nil
	}
}

// Equal returns a bool indicating if the two PathExps are equivalent.
func (pe *PathExp) Equal(other *PathExp) bool {

	switch {
	case pe.org != other.org:
		return false
	case pe.project != other.project:
		return false
	case !segmentsEqual(pe.envs, other.envs):
		return false
	case !segmentsEqual(pe.services, other.services):
		return false
	case !segmentsEqual(pe.identities, other.identities):
		return false
	case !segmentsEqual(pe.instances, other.instances):
		return false
	default:
		return true
	}
}

// CompareSpecificity returns an int indicating if this PathExp is more
// specific than PathExp b.
//
// PathExp A is more specific then PathExp B if, for each segment in the
// PathExp, A's segment is as specific or more specific than B's segment.
//
// Segment specificity is, from most to least specific:
//	- <literal>
//  - <glob>
//  - <alternation>
//  - <fullglob>
//
// It is assumed that the provided PathExps are not disjoint.
func (pe *PathExp) CompareSpecificity(other *PathExp) int {
	if cmp := compareSegmentType(pe.envs, other.envs); cmp != 0 {
		return cmp
	}

	if cmp := compareSegmentType(pe.services, other.services); cmp != 0 {
		return cmp
	}

	if cmp := compareSegmentType(pe.identities, other.identities); cmp != 0 {
		return cmp
	}

	return compareSegmentType(pe.instances, other.instances)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// This will be used in json decoding.
func (pe *PathExp) UnmarshalText(b []byte) error {
	o, err := Parse(string(b))
	if err != nil {
		return err
	}

	pe.org = o.org
	pe.project = o.project
	pe.envs = o.envs
	pe.services = o.services
	pe.instances = o.instances
	pe.identities = o.identities

	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
// This will be used in json encoding.
func (pe *PathExp) MarshalText() ([]byte, error) {
	return []byte(pe.String()), nil
}

// NewPartialFromPath creates a PathExp object from a cpath string that may be
// an incomplete path, resulting in added asterisks
func NewPartialFromPath(value string) (*PathExp, *int, error) {
	segments := strings.Split(value, "/")
	segments = segments[1:]
	if segments[len(segments)-1] == "" {
		segments = segments[:len(segments)-1]
	}

	org := stringAtIndex(segments, 0)
	project := stringAtIndex(segments, 1)

	envs, err := splitSliceAtIndex("environment", segments, 2)
	if err != nil {
		return nil, nil, errs.NewExitError(err.Error())
	}
	services, err := splitSliceAtIndex("service", segments, 3)
	if err != nil {
		return nil, nil, errs.NewExitError(err.Error())
	}
	users, err := splitSliceAtIndex("user", segments, 4)
	if err != nil {
		return nil, nil, errs.NewExitError(err.Error())
	}
	instances, err := splitSliceAtIndex("instance", segments, 5)
	if err != nil {
		return nil, nil, errs.NewExitError(err.Error())
	}

	return NewPartial(org, project, envs, services, users, instances)
}

// return string at index, or empty (prevent out of range)
func stringAtIndex(segments []string, idx int) string {
	if len(segments) > idx {
		return segments[idx]
	}
	return ""
}

// split cpathExp segment within []string at idx
func splitSliceAtIndex(name string, segments []string, idx int) ([]string, error) {
	if (len(segments) - 1) < idx {
		return []string{}, nil
	}
	return Split(name, segments[idx])
}
