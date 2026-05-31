package markitdown

import "fmt"

// Compile transforms a Markdown string into HTML.
//
// Example:
//
//	html, err := markitdown.Compile("# Hello World")
//	// html => "<h1>Hello World</h1><hr>"
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
