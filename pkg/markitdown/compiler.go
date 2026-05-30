package markitdown

func Compile(md string) string {
	tks := newLexer().tokenize(md)
	ast := newParser().parse(tks)
	out := newCodegen().gen(ast)
	return out
}
