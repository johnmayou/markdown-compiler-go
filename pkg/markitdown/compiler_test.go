package markitdown

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	result, err := Compile("")
	require.NoError(t, err)
	require.Equal(t, "", result)
}
