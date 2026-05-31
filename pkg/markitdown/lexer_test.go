package markitdown

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexerSmokeTest(t *testing.T) {
	got, err := newLexer("text").tokenize()
	require.NoError(t, err)

	want := []token{
		&textToken{text: "text", bold: false, italic: false},
		&newLineToken{},
	}

	require.Equal(t, want, got)
}
