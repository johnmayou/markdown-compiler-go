package markitdown

import "fmt"

func Compile(md string) (string, error) {
	tks, err := newLexer(md).tokenize()
	if err != nil {
		return "", fmt.Errorf("tokenize: %w", err)
	}
	ast, err := newParser(tks).parse()
	if err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}
	out, err := newCodegen(ast).gen()
	if err != nil {
		return "", fmt.Errorf("codegen: %w", err)
	}
	return out, nil
}
