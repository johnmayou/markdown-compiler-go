package markitdown

type token interface {
	isToken()
}

type headerToken struct {
	size int
}

func (t *headerToken) isToken() {}

type lexer struct{}

func newLexer() *lexer {
	return &lexer{}
}

func (l *lexer) tokenize(md string) []token {
	_ = md
	_ = headerToken{size: 0}
	return make([]token, 0)
}
