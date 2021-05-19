// Package gogen includes logic to generate files using Go template.
package gogen

import (
	"bytes"
	"errors"
	"go/ast"
	"go/types"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
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
)

// Description includes parsed information of a Go name. It will be used to feed data to the template.
type Description struct {
	Name    string
	Pkg     Package
	Methods []*Method
}

// Generator is an execution to generate code. Call Run method to trigger the job.
type Generator struct {
	Dir          string
	Name         string
	Output       string
	Template     string
	TemplateFile string
	Writer       io.Writer

	buf  *bytes.Buffer
	desc *Description
	pkgs []*packages.Package
}

// Method presents a method.
type Method struct {
	Name    string
	Params  []*Field
	Results []*Field
}

// Field presents a field.
type Field struct {
	Name string
	Type GoType
}

// GoType presents a Go type.
type GoType struct {
	types.Type
}

// Package presents a Go package.
type Package struct {
	Path string
	Name string
}

// Run executes the generator.
func (g *Generator) Run() error {
	log.Println("Generating code for", g.Name)

	steps := []func() error{
		g.parseSource,
		g.extractDescription,
		g.loadTemplate,
		g.executeTemplate,
		g.formatSource,
		g.writeToFile,
	}

	for _, f := range steps {
		if err := f(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) parseSource() error {
	log.Println("Parsing", g.Dir)

	pkgs, err := packages.Load(
		&packages.Config{
			Mode: parseMode,
		},
		g.Dir,
	)

	if err != nil {
		return err
	}

	g.pkgs = pkgs
	return nil
}

func (g *Generator) extractDescription() error {
	for _, pkg := range g.pkgs {
		for _, f := range pkg.Syntax {
			for _, decl := range f.Decls {
				if decl, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range decl.Specs {
						spec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}

						sType, ok := spec.Type.(*ast.InterfaceType)
						if !ok {
							continue
						}

						if spec.Name.Name != g.Name {
							continue
						}

						d := &Description{
							Name: g.Name,
							Pkg: Package{
								Path: pkg.PkgPath,
								Name: pkg.Name,
							},
							Methods: extractMethodsFromInterfaces(pkg, sType),
						}

						g.description = d
						return nil
					}
				}
			}
		}
	}

	return errors.New("gogen: no Go type found")
}

func (g *Generator) loadTemplate() error {
	if g.Template != "" {
		return nil
	}

	content, err := ioutil.ReadFile(g.TemplateFile)
	if err != nil {
		return err
	}

	g.template = string(templateContent)
	g.Template = string(content)
	return nil
}

func (g *Generator) executeTemplate() error {
	compiledTempl, err := template.New("gogen").
		Parse(g.Template)

	if err != nil {
		return err
	}

	g.buf = new(bytes.Buffer)
	return codeTmpl.Execute(g.buf, g.desc)
}

func (g *Generator) formatSource() error {
	out, err := imports.Process(g.OutFile, g.buf.Bytes(), nil)
	if err != nil {
		return err
	}

	g.buf = bytes.NewBuffer(out)
	return nil
}

func (g *Generator) writeToFile() error {
	f, err := os.Create(g.OutFile)
	if err != nil {
		return err
	}

	defer f.Close()

	_, errWrite := g.buf.WriteTo(f)
	return errWrite
}

func (g *Generator) writeToFile() error {
	if g.Writer != nil {
		_, err := g.buf.WriteTo(g.Writer)
		return err
	}

	if g.Output == "" {
		_, err := g.buf.WriteTo(os.Stdout)
		return err
	}

	return ioutil.WriteFile(g.Output, g.buf.Bytes(), fs.ModePerm)
}

func extractMethodsFromInterfaces(pkg *packages.Package, i *ast.InterfaceType) []*Method {
	var methods []*Method
	for _, method := range i.Methods.List {
		fnDesl, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		fn := &Method{
			Name:    method.Names[0].Name,
			Params:  extractFieldsFromAst(pkg, fnDesl.Params.List),
			Results: extractFieldsFromAst(pkg, fnDesl.Results.List),
		}

		methods = append(methods, fn)
	}

	return methods
}

func extractFieldsFromAst(pkg *packages.Package, items []*ast.Field) []*Field {
	output := []*Field{}

	for _, item := range items {
		name := ""

		//  nil if anonymous field
		if len(item.Names) > 0 {
			name = item.Names[0].Name
		}

		funcField := &Field{
			Type: GoType{
				Type: pkg.TypesInfo.TypeOf(item.Type),
			},
			Name: name,
		}

		output = append(output, funcField)
	}

	return output
}
