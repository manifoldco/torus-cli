/*
Package pathexp provides a representation of path expressions; locations of
secrets within the org/project/environment/service/identity/instance hierarchy
supporting globs and alternation.

Paths being parsed support <double-glob> but will be converted to <full-glob>
in the resulting pathexp

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
)

const slugstr = `[a-z\d][-_a-z\d]{0,63}`

var (
	slug           = regexp.MustCompile(`^` + slugstr + `$`)
	globRe         = regexp.MustCompile(`^(` + slugstr + `)(\*?)$`)
	fullglobOrGlob = regexp.MustCompile(`(^\*$)|(?:^(` + slugstr + `)(\*?)$)`)
	doubleGlob     = regexp.MustCompile(`^\*\*$`)
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
	Org        literal
	Project    literal
	Envs       segment
	Services   segment
	Identities segment
	Instances  segment
}

type segment interface {
	String() string
	Contains(subject string) bool
}

type literal string
type glob string
type fullglob struct{}
type alternation []segment

func (l literal) String() string { return string(l) }
func (l literal) Contains(subject string) bool {
	return string(l) == subject
}

func (g glob) String() string { return string(g) + "*" }
func (g glob) Contains(subject string) bool {
	return strings.Index(subject, string(g)) == 0
}

// GlobContains returns whether a glob, built from the value, contains the subject
func GlobContains(value, subject string) bool {
	gl := glob(value)
	return gl.Contains(subject)
}

func (f fullglob) String() string { return "*" }
func (f fullglob) Contains(subject string) bool {
	return true
}

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
func (a alternation) Contains(subject string) bool {
	for _, s := range a {
		if s.Contains(subject) {
			return true
		}
	}
	return false
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

// New creates a new path expression from the given path segments
// It returns an error if any of the values fail to validate
// and it must contain all relevant parts
func New(org, project string, envs, services, identities, instances []string) (*PathExp, error) {
	return newPathexp(org, project, envs, services, identities, instances, true)
}

// NewPartial creates a new path expression from the given path segments
// It returns an error if any of the values fail to validate
// but does not require any segments
func NewPartial(org, project string, envs, services, identities, instances []string) (*PathExp, error) {
	return newPathexp(org, project, envs, services, identities, instances, false)
}

// ValidSlug returns whether the subject is a valid slug value
func ValidSlug(subject string) bool {
	return slug.MatchString(subject)
}

func newPathexp(org, project string, envs, services, identities, instances []string, mustBeComplete bool) (*PathExp, error) {
	if org == "" {
		org = "*"
	}
	if project == "" {
		project = "*"
	}

	pe := PathExp{
		Org:     literal(org),
		Project: literal(project),
	}

	if !ValidSlug(org) && mustBeComplete {
		return nil, errors.New("invalid org name")
	}

	if !ValidSlug(project) && mustBeComplete {
		return nil, errors.New("invalid project name")
	}

	var err error

	if !mustBeComplete && len(envs) < 1 {
		return &pe, nil
	}
	pe.Envs, err = parseMultiple("environment", envs, mustBeComplete)
	if err != nil {
		return nil, err
	}

	if !mustBeComplete && len(services) < 1 {
		return &pe, nil
	}
	pe.Services, err = parseMultiple("service", services, mustBeComplete)
	if err != nil {
		return nil, err
	}

	if !mustBeComplete && len(identities) < 1 {
		return &pe, nil
	}
	pe.Identities, err = parseMultiple("identity", identities, mustBeComplete)
	if err != nil {
		return nil, err
	}

	if !mustBeComplete && len(instances) < 1 {
		return &pe, nil
	}
	pe.Instances, err = parseMultiple("instance", instances, mustBeComplete)
	if err != nil {
		return nil, err
	}

	return &pe, nil
}

// Parse parses a string into a path expression. It returns an error if the
// string is not a valid path expression.
func Parse(raw string) (*PathExp, error) {
	return parseStr(raw, true)
}

// ParsePartial parses a string into a path expression. It returns an error if the
// string is not a valid path expression, but allows missing segments
func ParsePartial(raw string) (*PathExp, error) {
	return parseStr(raw, false)
}

func parseStr(raw string, mustBeComplete bool) (*PathExp, error) {
	if len(raw) < 1 {
		return nil, errors.New("missing path segments")
	}

	parts := strings.Split(raw, "/")

	if mustBeComplete && len(parts) > 7 {
		return nil, errors.New("too many path segments")
	}
	if parts[0] != "" {
		return nil, errors.New("path expressions must start with '/'")
	}

	// remove leading empty section
	parts = parts[1:]

	// Identify if path contains doubleglob
	doubleGlobIndex, hasDoubleGlob, err := doubleGlobIndex(parts)
	if err != nil {
		return nil, err
	}

	// Replace parts with applicable wildcards if doubleglob exists
	segments := parts
	if hasDoubleGlob {
		segments = replaceGlobSegments(doubleGlobIndex, segments)
	}

	splitParts := make([][]string, 6)
	splitNames := []string{"", "", "environment", "service", "identity", "instance"}
	for i := 2; i < len(splitParts); i++ {
		segment := ""
		if len(segments) > i {
			segment = segments[i]
		}
		splitParts[i], err = Split(splitNames[i], segment)
		if err != nil {
			return nil, err
		}
	}

	org := ""
	if len(segments) > orgIdx {
		org = segments[orgIdx]
	}

	project := ""
	if len(segments) > projectIdx {
		project = segments[projectIdx]
	}

	if mustBeComplete {
		return New(
			org,
			project,
			splitParts[envIdx],
			splitParts[serviceIdx],
			splitParts[identityIdx],
			splitParts[instanceIdx],
		)
	}

	return NewPartial(
		org,
		project,
		splitParts[envIdx],
		splitParts[serviceIdx],
		splitParts[identityIdx],
		splitParts[instanceIdx],
	)
}

func doubleGlobIndex(parts []string) (int, bool, error) {
	var hasDoubleGlob bool
	var doubleGlobIndex int
	for i := range parts {
		if len(parts) > i && doubleGlob.MatchString(parts[i]) {
			if hasDoubleGlob {
				return doubleGlobIndex, hasDoubleGlob, errors.New("cannot use more than one **")
			}
			hasDoubleGlob = true
			doubleGlobIndex = i
		}
	}
	return doubleGlobIndex, hasDoubleGlob, nil
}

func replaceGlobSegments(doubleGlobIndex int, parts []string) []string {
	segments := []string{"", "", "*", "*", "*", "*"}
	leading := parts[:doubleGlobIndex]
	copy(segments, leading)
	trailing := parts[doubleGlobIndex+1:]
	for i, part := range trailing {
		idx := len(segments) - len(trailing) + i
		segments[idx] = part
	}

	return segments
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

	segment, err := parseMultiple("instance", parts, true)
	if err != nil {
		return nil, err
	}

	return &PathExp{
		Org:        pe.Org,
		Project:    pe.Project,
		Envs:       pe.Envs,
		Services:   pe.Services,
		Identities: pe.Identities,
		Instances:  segment,
	}, nil
}

// String returns the unparsed string representation of the path expression
func (pe *PathExp) String() string {
	return strings.Join([]string{"", string(pe.Org), string(pe.Project),
		pe.Envs.String(),
		pe.Services.String(),
		pe.Identities.String(),
		pe.Instances.String(),
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

func parseMultiple(name string, parts []string, mustBeComplete bool) (segment, error) {
	switch len(parts) {
	case 0:
		return nil, errors.New("Empty segment alternation for " + name + ".")
	case 1:
		matches := fullglobOrGlob.FindAllStringSubmatch(parts[0], -1)
		if len(matches) != 1 {
			if mustBeComplete {
				return nil, errors.New("Invalid " + name + ".")
			}
			return fullglob{}, nil
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
	case pe.Org != other.Org:
		return false
	case pe.Project != other.Project:
		return false
	case !segmentsEqual(pe.Envs, other.Envs):
		return false
	case !segmentsEqual(pe.Services, other.Services):
		return false
	case !segmentsEqual(pe.Identities, other.Identities):
		return false
	case !segmentsEqual(pe.Instances, other.Instances):
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
	if cmp := compareSegmentType(pe.Envs, other.Envs); cmp != 0 {
		return cmp
	}

	if cmp := compareSegmentType(pe.Services, other.Services); cmp != 0 {
		return cmp
	}

	if cmp := compareSegmentType(pe.Identities, other.Identities); cmp != 0 {
		return cmp
	}

	return compareSegmentType(pe.Instances, other.Instances)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// This will be used in json decoding.
func (pe *PathExp) UnmarshalText(b []byte) error {
	o, err := Parse(string(b))
	if err != nil {
		return err
	}

	pe.Org = o.Org
	pe.Project = o.Project
	pe.Envs = o.Envs
	pe.Services = o.Services
	pe.Instances = o.Instances
	pe.Identities = o.Identities

	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
// This will be used in json encoding.
func (pe *PathExp) MarshalText() ([]byte, error) {
	return []byte(pe.String()), nil
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
