package markitdown

type token interface {
	kind()
}

type lexer struct{}

func newLexer() *lexer {
	return &lexer{}
}

func (l *lexer) tokenize(md string) []token {
	_ = md
	return make([]token, 0)
}
