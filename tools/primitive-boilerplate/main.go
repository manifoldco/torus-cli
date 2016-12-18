package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// Our output files
const (
	primitiveFile = "primitive/zz_generated_primitive.go"
	envelopeFile  = "envelope/zz_generated_envelope.go"
	cryptoFile    = "daemon/crypto/zz_generated_crypto.go"
)

var primitiveTmpl, envelopeTmpl, cryptoTmpl *template.Template

func init() {
	prefix := "tools/primitive-boilerplate/"
	primitiveName := filepath.Join(prefix, "primitive.tmpl")
	primitiveTmpl = template.Must(template.ParseFiles(primitiveName))

	envelopeName := filepath.Join(prefix, "envelope.tmpl")
	envelopeTmpl = template.Must(template.New("envelope.tmpl").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}).ParseFiles(envelopeName))

	cryptoName := filepath.Join(prefix, "crypto.tmpl")
	cryptoTmpl = template.Must(template.ParseFiles(cryptoName))
}

// data is a representation of one of our primitive structs, or a struct
// referenced by one of our primitives.
type data struct {
	Name      string
	Fields    []field
	Byte      string // The enumerated byte type, if present.
	Version   uint8  // schema version, if present.
	Immutable bool   // Is the type immutable?
	Visible   bool   // Is the type visible from an immutable type?
}

// field is a representation of a struct field.
type field struct {
	Name  string // The JSON representation
	Field string // The actual struct field name
}

// sortData and sortFields are used for sorting the data and field types.
// data is sorted by the enumerated byte type. fields are sorted by name.

type sortData []data

func (d sortData) Len() int           { return len(d) }
func (d sortData) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d sortData) Less(i, j int) bool { return strings.Compare(d[i].Byte, d[j].Byte) < 0 }

type sortFields []field

func (s sortFields) Len() int           { return len(s) }
func (s sortFields) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortFields) Less(i, j int) bool { return strings.Compare(s[i].Name, s[j].Name) < 0 }

// sortDataByVersion sorts data by schema version, for envelope output.
type sortDataByVersion []data

func (d sortDataByVersion) Len() int           { return len(d) }
func (d sortDataByVersion) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d sortDataByVersion) Less(i, j int) bool { return d[i].Version < d[j].Version }

// envTmplData is the data passed for populating the envelope package template
// file, defined here for convenience.
type envTmplData struct {
	Name       string
	Mutability string
	Types      map[string][]data
}

// primitiveParser handles iterating and inspecting the defined primitive structs
type primitiveParser struct {
	fset *token.FileSet
	p    *ast.File

	typedecls map[int]string // our comment annotated type bytes
	structmap map[string]*ast.StructType
}

// Parse performs the actual work of walking over all defined structs,
// and building data representations of them for use in template rendering.
func (pp *primitiveParser) Parse() map[string]*data {
	pp.loadTypeComments()
	pp.loadStructs()

	visible := make(map[string]*data)
	for k := range pp.structmap {
		if strings.ToUpper(string(k[0])) != string(k[0]) { // not exported
			continue
		}
		pp.visitStruct(visible, k, false)
	}

	return visible
}

// visitStruct inspects the struct with the given name, parsing out its
// field definitions and mutability state.
func (pp *primitiveParser) visitStruct(reachable map[string]*data, k string, visible bool) {
	if _, ok := reachable[k]; ok { // already included
		return
	}

	s := pp.structmap[k]

	d := data{
		Name: k,
		Byte: pp.typedecls[pp.fset.Position(s.Pos()).Line],
	}
	d.Immutable, d.Version, d.Fields = pp.gatherFields(reachable, s, visible)
	d.Visible = visible || d.Immutable

	sort.Sort(sortFields(d.Fields))
	reachable[k] = &d
}

// gatherFields is used by visitStruct to gather the fields for a struct.
// embedded struct fields are hoisted up to the enclosing struct for JSON
// representation.
// referenced structs will in turn be visited.
//
// gatherFields transfers visibility from immutable structs, or any already
// known visible structs, to structs that they reference.
// visibility determines if we should create a MarshalJSON func for the struct.
func (pp *primitiveParser) gatherFields(reachable map[string]*data, s *ast.StructType, visible bool) (bool, uint8, []field) {
	fields := []field{}

	immutable := false
	var version uint8
	if s == nil {
		return immutable, version, fields
	}

	for _, f := range s.Fields.List {
		var typeName string
		switch t := f.Type.(type) {
		case *ast.Ident:
			typeName = t.Name
		case *ast.StarExpr:
			if i, ok := t.X.(*ast.Ident); ok {
				typeName = i.Name
			}
		}
		if len(f.Names) == 0 {
			switch typeName {
			case "immutable":
				immutable = true
			case "v1Schema":
				version = 1
			case "v2Schema":
				version = 2
			}
			embedded := pp.structmap[typeName]
			embImmutable, embVersion, embeddedFields := pp.gatherFields(reachable, embedded, visible || immutable)
			immutable = immutable || embImmutable
			visible = visible || immutable
			if version == 0 {
				version = embVersion
			}
			fields = append(fields, embeddedFields...)
			continue
		}

		// This isn't an embedded struct, but it might be our own struct
		// that is now reachable. we need to create a MarshalJSON func for
		// it if so.
		if _, ok := pp.structmap[typeName]; ok {
			pp.visitStruct(reachable, typeName, visible)
		}

		for _, n := range f.Names {
			name := n.Name
			if f.Tag != nil {
				tname := parseJSONTag(f.Tag.Value)
				if tname != "" {
					name = tname
				}
			}
			fields = append(fields, field{
				Name:  name,
				Field: n.Name,
			})
		}
	}

	return immutable, version, fields
}

// loadTypeComments loads our annotation comments on structs, indicating
// their enumerated byte type, allowing us to look it up later by line.
func (pp *primitiveParser) loadTypeComments() {
	pp.typedecls = make(map[int]string)
	for _, c := range pp.p.Comments {
		cmt := strings.Trim(c.List[0].Text, "/* \t")
		parts := strings.Split(cmt, " ")
		if len(parts) == 2 && parts[0] == "type:" {
			pp.typedecls[pp.fset.Position(c.List[0].Pos()).Line] = parts[1]
		}
	}
}

// loadStructs creates a map of name to ast node for struct types from the
// parsed file, filtering out other types.
func (pp *primitiveParser) loadStructs() {
	pp.structmap = make(map[string]*ast.StructType)

	for k, o := range pp.p.Scope.Objects {
		// We only care about types
		if o.Kind != ast.Typ {
			continue
		}
		ts, ok := o.Decl.(*ast.TypeSpec)
		if !ok {
			continue
		}

		// and to be more exact, only struct types
		s, ok := ts.Type.(*ast.StructType)
		if !ok {
			continue
		}

		// Empty struct? We don't need to generate and json for it. skip.
		if len(s.Fields.List) == 0 {
			continue
		}

		pp.structmap[k] = s
	}
}

// parseJSONTag parses the given JSON struct field tag, to get its name.
// omitting fields via tags is not supported.
func parseJSONTag(tag string) string {
	if len(tag) == 0 {
		return ""
	}

	if tag[:7] != "`json:\"" {
		return ""
	}

	parts := strings.Split(tag[7:len(tag)-2], ",")

	return parts[0]
}

// writeTemplate writes the given template to the given file, with the given
// data. Templates are assumed to be generating go source files, so they
// are formatted as well.
func writeTemplate(fileName string, tmpl *template.Template, data interface{}) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		log.Fatal(err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	fd, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	fd.Write(formatted)
}

func main() {
	fset := token.NewFileSet()
	p, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	pp := &primitiveParser{fset: fset, p: p}
	visible := pp.Parse()

	// Sort all types by byte type, for primitive file output.
	var types []data
	for _, d := range visible {
		types = append(types, *d)
	}
	sort.Sort(sortData(types))

	// Split the sorted values into immutable/mutable, for envelope output.
	mutable := map[string][]data{}
	immutable := map[string][]data{}
	for _, d := range types {
		if d.Byte == "" {
			continue
		}

		if d.Immutable {
			immutable[d.Byte] = append(immutable[d.Byte], d)
		} else {
			mutable[d.Byte] = append(mutable[d.Byte], d)
		}
	}

	// Template output will automatically sort maps for us, but sorted schema
	// versions in the switch statements would be nice, too.
	for _, d := range mutable {
		sort.Sort(sortDataByVersion(d))
	}
	for _, d := range immutable {
		sort.Sort(sortDataByVersion(d))
	}

	writeTemplate(primitiveFile, primitiveTmpl, struct{ Types []data }{Types: types})

	writeTemplate(envelopeFile, envelopeTmpl, []envTmplData{
		{Name: "Unsigned", Mutability: "Mutable", Types: mutable},
		{Name: "Signed", Mutability: "Immutable", Types: immutable},
	})

	writeTemplate(cryptoFile, cryptoTmpl, immutable)
}
