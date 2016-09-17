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

	"github.com/kardianos/osext"
)

var primitiveTmpl = template.Must(func() (*template.Template, error) {
	d, _ := osext.ExecutableFolder()
	return template.ParseFiles(filepath.Join(d, "../primitive-boilerplate/primitive.tmpl"))
}())

var envelopeTmpl = template.Must(func() (*template.Template, error) {
	d, _ := osext.ExecutableFolder()
	return template.ParseFiles(filepath.Join(d, "../primitive-boilerplate/envelope.tmpl"))
}())

type envTmplData struct {
	Name       string
	Mutability string
	Types      []data
}

type field struct {
	Name  string
	Field string
}

type sortFields []field

func (s sortFields) Len() int           { return len(s) }
func (s sortFields) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortFields) Less(i, j int) bool { return strings.Compare(s[i].Name, s[j].Name) < 0 }

type data struct {
	Name      string
	Fields    []field
	Byte      string
	Immutable bool
	Visible   bool
}

type sortData []data

func (d sortData) Len() int           { return len(d) }
func (d sortData) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d sortData) Less(i, j int) bool { return strings.Compare(d[i].Byte, d[j].Byte) < 0 }

func main() {
	fset := token.NewFileSet()
	p, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	typedecls := make(map[int]string)
	for _, c := range p.Comments {
		cmt := strings.Trim(c.List[0].Text, "/* \t")
		parts := strings.Split(cmt, " ")
		if len(parts) == 2 && parts[0] == "type:" {
			typedecls[fset.Position(c.List[0].Pos()).Line] = parts[1]
		}
	}

	var types []data

	structmap := make(map[string]*ast.StructType)

	for k, o := range p.Scope.Objects {
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

		structmap[k] = s
	}

	reachable := make(map[string]*data)
	for k := range structmap {
		if strings.ToUpper(string(k[0])) != string(k[0]) { // not exported
			continue
		}
		reachStruct(fset, typedecls, reachable, structmap, k, false)
	}

	for _, d := range reachable {
		types = append(types, *d)
	}
	sort.Sort(sortData(types))

	mutable := []data{}
	immutable := []data{}
	for _, d := range types {
		if d.Byte == "" {
			continue
		}

		if d.Immutable {
			immutable = append(immutable, d)
		} else {
			mutable = append(mutable, d)
		}
	}

	writeTemplate("primitive/zz_generated_primitive.go", primitiveTmpl,
		struct{ Types []data }{Types: types})

	writeTemplate("envelope/zz_generated_envelope.go", envelopeTmpl, []envTmplData{
		{Name: "Unsigned", Mutability: "Mutable", Types: mutable},
		{Name: "Signed", Mutability: "Immutable", Types: immutable},
	})
}

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

func reachStruct(fset *token.FileSet, typedecls map[int]string, reachable map[string]*data, structmap map[string]*ast.StructType, k string, visible bool) {
	if _, ok := reachable[k]; ok { // already included
		return
	}

	s := structmap[k]

	d := data{
		Name: k,
		Byte: typedecls[fset.Position(s.Pos()).Line],
	}
	d.Immutable, d.Fields = gatherFields(fset, typedecls, reachable, structmap, s, visible)
	d.Visible = visible || d.Immutable

	sort.Sort(sortFields(d.Fields))
	reachable[k] = &d
}

func gatherFields(fset *token.FileSet, typedecls map[int]string, reachable map[string]*data, structmap map[string]*ast.StructType, s *ast.StructType, visible bool) (bool, []field) {
	fields := []field{}

	immutable := false
	if s == nil {
		return immutable, fields
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
			if typeName == "immutable" {
				immutable = true
			}
			embedded := structmap[typeName]
			embImmutable, embeddedFields := gatherFields(fset, typedecls, reachable, structmap, embedded, visible || immutable)
			immutable = immutable || embImmutable
			visible = visible || immutable
			fields = append(fields, embeddedFields...)
			continue
		}

		// This isn't an embedded struct, but it might be our own struct
		// that is now reachable. we need to create a MarshalJSON func for
		// it if so.
		if _, ok := structmap[typeName]; ok {
			reachStruct(fset, typedecls, reachable, structmap, typeName, visible)
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

	return immutable, fields
}

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
