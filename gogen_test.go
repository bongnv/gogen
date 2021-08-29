package gogen

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Generator_Run_empty(t *testing.T) {
	buf := new(bytes.Buffer)
	g := &Generator{
		Name:     "Mock",
		Template: "{{ .Name }} is generated.",
		Writer:   buf,
	}

	err := g.Run()
	require.NoError(t, err)
	require.Equal(t, "Mock is generated.", buf.String())
}

func Test_Generator_Run_parse_interface(t *testing.T) {
	ctx := context.Background()

	wd, err := os.Getwd()
	require.NoError(t, err)

	g := &Generator{
		Dir:  filepath.Join(wd, "examples", "noop"),
		Name: "Example",
	}

	require.NoError(t, g.parseSource(ctx))
	require.NotNil(t, g.pkg)

	require.NoError(t, g.prepareDescription(ctx))
	require.NotNil(t, g.desc)
	require.NotNil(t, g.desc.Pkg, "package information must be available")
	require.Equal(t, "noop", g.desc.Pkg.Name)
	require.Equal(t, "github.com/bongnv/gogen/examples/noop", g.desc.Pkg.Path)

	require.NoError(t, g.parseImports(ctx))
	require.Len(t, g.desc.Imports, 1)
	require.Equal(t, "context", g.desc.Imports[0].Name)
	require.Equal(t, "context", g.desc.Imports[0].Path)

	require.NoError(t, g.parseTypeInfo(ctx))
	require.True(t, g.desc.IsInterface)
	require.Len(t, g.desc.Methods, 1)
	initMethod := g.desc.Methods[0]
	require.Equal(t, "Init", initMethod.Name)
	require.Len(t, initMethod.Params, 1)
	require.Equal(t, "ctx", initMethod.Params[0].Name)
	require.Equal(t, "context.Context", initMethod.Params[0].Type.String())
	require.Len(t, initMethod.Results, 1)
	require.Equal(t, "", initMethod.Results[0].Name)
	require.Equal(t, "error", initMethod.Results[0].Type.String())
}

func Test_Generator_Run_parse_struct(t *testing.T) {
	ctx := context.Background()

	wd, err := os.Getwd()
	require.NoError(t, err)

	g := &Generator{
		Dir:  filepath.Join(wd, "examples", "getter"),
		Name: "Example",
	}

	require.NoError(t, g.parseSource(ctx))
	require.NotNil(t, g.pkg)

	require.NoError(t, g.prepareDescription(ctx))
	require.NotNil(t, g.desc)
	require.NotNil(t, g.desc.Pkg, "package information must be available")
	require.Equal(t, "getter", g.desc.Pkg.Name)

	require.NoError(t, g.parseTypeInfo(ctx))
	require.True(t, g.desc.IsStruct)
	require.Len(t, g.desc.Fields, 3)
	fields := g.desc.Fields
	require.Equal(t, "Number", fields[0].Name)
	require.Equal(t, "int", fields[0].Type.String())
	require.Equal(t, "String", fields[1].Name)
	require.Equal(t, "string", fields[1].Type.String())
	require.Equal(t, "StringPtr", fields[2].Name)
	require.Equal(t, "*string", fields[2].Type.String())
	require.True(t, fields[2].Type.IsPointer)
	require.Equal(t, "true", fields[2].Tags["skip"])
}

func Test_formatSource(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	output := filepath.Join(wd, "mock_file.go")
	content := `
	package gogen

	type Service interface {}`

	g := &Generator{
		Output: output,
		Format: true,
		buf:    bytes.NewBufferString(content),
	}
	require.NoError(t, g.formatSource(context.Background()))
	require.Equal(t, "package gogen\n\ntype Service interface{}\n", g.buf.String(), "content must be formatted")
}
