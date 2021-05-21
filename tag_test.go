package gogen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseTagItems(t *testing.T) {
	tags, err := parseTagItems("skip")
	require.NoError(t, err)
	require.Contains(t, tags, "skip")
}

func Test_extractTags(t *testing.T) {
	tags, err := extractTags(`gogen:"skip"`)
	require.NoError(t, err)
	require.Equal(t, "true", tags["skip"])
}
