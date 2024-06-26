package parser

import (
	"testing"

	"github.com/reedom/convergen/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoConvergenInterface(t *testing.T) {
	c, err := NewParser(
		&config.Config{
			Input:  "../../tests/fixtures/usecase/nointf/setup.go",
			Output: "../../tests/fixtures/usecase/nointf/setup.gen.go",
		},
	)
	require.Nil(t, err)
	_, err = c.findConvergenEntries()
	require.NotNil(t, err)
	assert.ErrorContains(t, err, "nointf/setup.go:1:1: Convergen interface not found")
}
