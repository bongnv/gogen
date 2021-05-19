package main

import (
	"github.com/alecthomas/kong"
	"github.com/bongnv/gogen"
)

var cli struct {
	Name     string `kong:"arg,required,help='Name of a Go type'"`
	Template string `kong:"required,type='existingfile',short='t',help='Path to the template file'"`
	Output   string `kong:"help='Path to the output',short='o'"`
	Dir      string `kong:"short='d',help='Directory to search for',default='.'"`
	Format   bool   `kong:"short='f',help='Format source code'"`
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.Name("gogen"),
		kong.Description("A code generation tool using Go template"),
	)

	g := &gogen.Generator{
		Dir:          cli.Dir,
		Name:         cli.Name,
		TemplateFile: cli.Template,
		Output:       cli.Output,
		Format:       cli.Format,
	}

	ctx.FatalIfErrorf(g.Run())
}
