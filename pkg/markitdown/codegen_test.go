package markitdown

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodegen(t *testing.T) {
	got, err := newCodegen(
		&astRootNode{
			children: []astNode{
				&astParagraphNode{children: []astInlineNode{&astTextNode{text: "text", bold: false, italic: false}}},
			},
		},
	).gen()
	require.NoError(t, err)

	require.Equal(t, "<p>text</p>", got)
}
