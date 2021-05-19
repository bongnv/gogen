package main

import (
	"github.com/alecthomas/kong"
	"github.com/bongnv/gogen"
)

var cli struct {
	Name     string `kong:"arg,required,help='Name of a Go type'"`
	Template string `kong:"required,type='existingfile',short='t',help='Path to the template file'"`
	Output   string `kong:"help='Path to the output',short='o'"`
}

func main() {
	ctx := kong.Parse(
		&cli,
		kong.Name("gogen"),
		kong.Description("A code generation tool using Go template"),
	)

	g := &gogen.Generator{
		Name:         cli.Name,
		TemplateFile: cli.Template,
		Output:       cli.Output,
	}

	ctx.FatalIfErrorf(g.Run())
}
