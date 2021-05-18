package main

import (
	"flag"
	"log"
	"os"

	"github.com/bongnv/gogen"
)

func main() {
	g := &gogen.Generator{}
	flags := flag.NewFlagSet("gogen", flag.ExitOnError)
	flags.BoolVar(&g.Debug, "debug", false, "Enable debugging")
	flags.StringVar(&g.Dir, "dir", ".", "Path to the source code directly")
	flags.StringVar(&g.Dir, "name", "", "Name of the Go type to be parsed")
	flags.StringVar(&g.TemplateFile, "tempFile", "", "Path to the template file")
	flags.StringVar(&g.OutFile, "outFile", "", "Path to the output file")
	flags.Parse(os.Args[1:])

	if err := g.Run(); err != nil {
		log.Fatalln(err)
	}
}
