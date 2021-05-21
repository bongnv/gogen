package gogen

import (
	"bytes"
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

func Test_Generator_Run_with_source(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	g := &Generator{
		Dir:  filepath.Join(wd, "examples", "noop"),
		Name: "Example",
	}

	require.NoError(t, g.parseSource())
	require.NotNil(t, g.pkg)

	require.NoError(t, g.prepareDescription())
	require.NotNil(t, g.desc)
	require.NotNil(t, g.desc.Pkg, "package information must be available")
	require.Equal(t, "noop", g.desc.Pkg.Name)
	require.Equal(t, "github.com/bongnv/gogen/examples/noop", g.desc.Pkg.Path)

	require.NoError(t, g.parseImports())
	require.Len(t, g.desc.Imports, 1)
	require.Equal(t, "context", g.desc.Imports[0].Name)
	require.Equal(t, "context", g.desc.Imports[0].Path)

	require.NoError(t, g.parseTypeInfo())
	require.True(t, g.desc.IsInterface)
	require.Len(t, g.desc.Methods, 1)
	initMethod := g.desc.Methods[0]
	require.Equal(t, "Init", initMethod.Name)
	require.Len(t, initMethod.Params, 1)
	require.Equal(t, "ctx", initMethod.Params[0].Name)
	require.Equal(t, "context.Context", initMethod.Params[0].Type)
	require.Len(t, initMethod.Results, 1)
	require.Equal(t, "", initMethod.Results[0].Name)
	require.Equal(t, "error", initMethod.Results[0].Type)
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
	require.NoError(t, g.formatSource())
	require.Equal(t, "package gogen\n\ntype Service interface{}\n", g.buf.String(), "content must be formatted")
}
