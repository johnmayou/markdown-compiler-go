package markitdown

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	got, err := newParser(
		[]token{
			&textToken{text: "text", bold: false, italic: false},
			&newLineToken{},
		},
	).parse()
	require.NoError(t, err)

	want := &astRootNode{
		children: []astNode{
			&astParagraphNode{children: []astInlineNode{&astTextNode{text: "text", bold: false, italic: false}}},
		},
	}

	require.Equal(t, want, got)
}
