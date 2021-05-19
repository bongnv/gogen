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
		Dir:  wd,
		Name: "Generator",
	}

	require.NoError(t, g.parseSource())
	require.NotEmpty(t, g.pkgs)
	require.NoError(t, g.extractDescription())
	require.NotNil(t, g.desc)
	require.NotNil(t, g.desc.Pkg, "package information must be available")
	require.Equal(t, "gogen", g.desc.Pkg.Name)
	require.Equal(t, "github.com/bongnv/gogen", g.desc.Pkg.Path)
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
