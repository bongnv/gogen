// Package gogen includes logic to generate files using Go template.
package gogen

import (
	"bytes"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
)

type Description struct {
	Name string
}

type Generator struct {
	Name         string
	Output       string
	TemplateFile string
	Writer       io.Writer

	buf      *bytes.Buffer
	desc     *Description
	template string
}

func (g *Generator) Run() error {
	log.Println("Generating code for", g.Name)

	steps := []func() error{
		g.extractDescription,
		g.loadTemplate,
		g.executeTemplate,
		g.writeToFile,
	}

	for _, f := range steps {
		if err := f(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) extractDescription() error {
	g.desc = &Description{
		Name: g.Name,
	}
	return nil
}

func (g *Generator) loadTemplate() error {
	content, err := ioutil.ReadFile(g.TemplateFile)
	if err != nil {
		return err
	}

	g.template = string(content)
	return nil
}

func (g *Generator) executeTemplate() error {
	compiledTempl, err := template.New("gogen").
		Parse(g.template)

	if err != nil {
		return err
	}

	g.buf = new(bytes.Buffer)
	return compiledTempl.Execute(g.buf, g.desc)
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
