package gogen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Generator_Run(t *testing.T) {
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
