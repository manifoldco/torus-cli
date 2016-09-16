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

type field struct {
	Name  string
	Field string
}

type sortFields []field

func (s sortFields) Len() int           { return len(s) }
func (s sortFields) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortFields) Less(i, j int) bool { return strings.Compare(s[i].Name, s[j].Name) < 0 }

type data struct {
	Name   string
	Fields []field
	Byte   string
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

		// Empty struct? likely one ofthe schema types. skip.
		if len(s.Fields.List) == 0 {
			continue
		}

		d := data{
			Name: k,
			Byte: typedecls[fset.Position(s.Pos()).Line],
		}

		for _, f := range s.Fields.List {

			if len(f.Names) == 0 {
				// embedded struct
			} else {
				for _, n := range f.Names {
					name := n.Name
					if f.Tag != nil {
						tname := parseJSONTag(f.Tag.Value)
						if tname != "" {
							name = tname
						}
					}
					d.Fields = append(d.Fields, field{
						Name:  name,
						Field: n.Name,
					})
				}
			}
		}

		sort.Sort(sortFields(d.Fields))
		types = append(types, d)
	}

	sort.Sort(sortData(types))
	tmplData := struct {
		Types []data
	}{
		Types: types,
	}

	writeTemplate("primitive/zz_generated_primitive.go", primitiveTmpl, tmplData)
	writeTemplate("envelope/zz_generated_envelope.go", envelopeTmpl, tmplData)
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
