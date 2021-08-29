// Package gogen includes logic to generate files using Go template.
package gogen

import (
	"bytes"
	"context"
	"go/ast"
	"go/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/bongnv/task"
	"golang.org/x/tools/go/packages"
	goimports "golang.org/x/tools/imports"
)

var (
	parseMode = packages.NeedName |
		packages.NeedFiles |
		packages.NeedImports |
		packages.NeedDeps |
		packages.NeedCompiledGoFiles |
		packages.NeedTypes |
		packages.NeedSyntax |
		packages.NeedTypesInfo

	funcsMap template.FuncMap = template.FuncMap{
		"Quote": strconv.Quote,
	}
)

// Description includes parsed information of a Go name. It will be used to feed data to the template.
type Description struct {
	Name        string
	Pkg         *Package
	Imports     []*Package
	IsInterface bool
	IsStruct    bool
	Methods     []*Method
	Fields      []*Field
}

// GoType presents a type in Go
type GoType struct {
	Name      string
	IsPointer bool
}

// String returns the string presentation.
func (t GoType) String() string {
	return t.Name
}

// Var is a pair of name and Go type to present a variable.
type Var struct {
	Name string
	Type *GoType
}

// Field is a field in a struct.
type Field struct {
	Name string
	Type *GoType
	Tags map[string]string
}

// Method defines a method in Go.
type Method struct {
	Name    string
	Params  []*Var
	Results []*Var
}

// Generator is an execution to generate code. Call Run method to trigger the job.
type Generator struct {
	Dir          string
	Format       bool
	Name         string
	Output       string
	Template     string
	TemplateFile string
	Writer       io.Writer

	buf      *bytes.Buffer
	desc     *Description
	pkg      *packages.Package
	typeInfo types.Type
	typeSpec *ast.TypeSpec
}

// Package presents a Go package.
type Package struct {
	Path string
	Name string
}

// Run executes the generator.
func (g *Generator) Run() error {
	log.Println("Generating code for", g.Name)

	return task.Exec(
		context.Background(),
		g.parseSource,
		g.prepareDescription,
		g.parseImports,
		g.parseTypeInfo,
		g.loadTemplate,
		g.executeTemplate,
		g.formatSource,
		g.writeToFile,
	)
}

func (g *Generator) parseSource(ctx context.Context) error {
	dir, err := filepath.Abs(g.Dir)
	if err != nil {
		return err
	}

	log.Println("Parsing", dir)

	pkgs, err := packages.Load(
		&packages.Config{
			Mode: parseMode,
		},
		dir,
	)

	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			for _, decl := range f.Decls {
				if decl, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range decl.Specs {
						spec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}

						if spec.Name.Name != g.Name {
							continue
						}

						g.pkg = pkg
						g.typeInfo = pkg.TypesInfo.TypeOf(spec.Type)
						g.typeSpec = spec

						return nil
					}
				}
			}
		}
	}

	return nil
}

func (g *Generator) prepareDescription(ctx context.Context) error {
	g.desc = &Description{
		Name: g.Name,
	}

	if g.pkg == nil {
		return nil
	}

	g.desc.Pkg = &Package{
		Path: g.pkg.PkgPath,
		Name: g.pkg.Name,
	}

	return nil
}

func (g *Generator) parseImports(ctx context.Context) error {
	if g.pkg == nil || len(g.pkg.Imports) == 0 {
		return nil
	}

	// TODO: support duplicate name
	for _, p := range g.pkg.Imports {
		g.desc.Imports = append(g.desc.Imports, &Package{
			Name: p.Name,
			Path: p.PkgPath,
		})
	}

	return nil
}

func (g *Generator) loadTemplate(ctx context.Context) error {
	if g.Template != "" {
		return nil
	}

	content, err := ioutil.ReadFile(g.TemplateFile)
	if err != nil {
		return err
	}

	g.Template = string(content)
	return nil
}

func (g *Generator) executeTemplate(ctx context.Context) error {
	compiledTempl, err := template.New("gogen").
		Funcs(funcsMap).
		Parse(g.Template)

	if err != nil {
		return err
	}

	g.buf = new(bytes.Buffer)
	return compiledTempl.Execute(g.buf, g.desc)
}

func (g *Generator) formatSource(ctx context.Context) error {
	if !g.Format {
		return nil
	}

	out, err := goimports.Process(g.Output, g.buf.Bytes(), nil)
	if err != nil {
		return err
	}

	g.buf = bytes.NewBuffer(out)
	return nil
}

func (g *Generator) writeToFile(ctx context.Context) error {
	if g.Writer != nil {
		_, err := g.buf.WriteTo(g.Writer)
		return err
	}

	if g.Output == "" {
		_, err := g.buf.WriteTo(os.Stdout)
		return err
	}

	return ioutil.WriteFile(g.Output, g.buf.Bytes(), 0644)
}

func (g *Generator) parseTypeInfo(ctx context.Context) error {
	if g.typeInfo == nil {
		return nil
	}

	switch v := g.typeInfo.(type) {
	case *types.Interface:
		g.desc.IsInterface = true
		g.desc.Methods = extractMethods(v)
	case *types.Struct:
		g.desc.IsStruct = true
		fields, err := extractFields(v)
		if err != nil {
			return err
		}
		g.desc.Fields = fields
	}

	return nil
}

func extractMethods(typeInfo *types.Interface) []*Method {
	methods := make([]*Method, typeInfo.NumExplicitMethods())
	for i := 0; i < typeInfo.NumExplicitMethods(); i++ {
		fn := typeInfo.ExplicitMethod(i)
		signature := fn.Type().(*types.Signature)
		methods[i] = &Method{
			Name:    fn.Name(),
			Params:  extractVariables(signature.Params()),
			Results: extractVariables(signature.Results()),
		}
	}
	return methods
}

func extractVariables(tuple *types.Tuple) []*Var {
	vars := make([]*Var, tuple.Len())
	for i := tuple.Len() - 1; i >= 0; i-- {
		currentVar := tuple.At(i)
		vars[i] = &Var{
			Name: currentVar.Name(),
			Type: extractGoType(currentVar.Type()),
		}
	}

	return vars
}

func extractFields(typeInfo *types.Struct) ([]*Field, error) {
	fields := make([]*Field, typeInfo.NumFields())

	for i := typeInfo.NumFields() - 1; i >= 0; i-- {
		field := typeInfo.Field(i)
		tags, err := extractTags(typeInfo.Tag(i))
		if err != nil {
			return nil, err
		}

		fields[i] = &Field{
			Name: field.Name(),
			Type: extractGoType(field.Type()),
			Tags: tags,
		}
	}

	return fields, nil
}

// TODO: implements cache to speed up
func extractGoType(typeInfo types.Type) *GoType {
	_, isPointer := typeInfo.(*types.Pointer)
	return &GoType{
		Name:      typeInfo.String(),
		IsPointer: isPointer,
	}
}
