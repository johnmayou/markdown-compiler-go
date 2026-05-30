package markitdown

import (
	"fmt"
	"strings"
)

type codegen struct {
	ast  *astRootNode
	html []string
}

// newCodegen returns a codegen initialized with the given AST root.
func newCodegen(ast *astRootNode) *codegen {
	return &codegen{ast: ast}
}

// gen walks the AST root and emits a complete HTML string, returning an error for unrecognized node types.
func (c *codegen) gen() (string, error) {
	for _, node := range c.ast.children {
		switch node := node.(type) {
		case *astHeaderNode:
			c.html = append(c.html, c.genHeader(node))
		case *astCodeBlockNode:
			c.html = append(c.html, c.genCodeBlock(node))
		case *astQuoteNode:
			c.html = append(c.html, c.genQuoteBlock(node))
		case *astListNode:
			c.html = append(c.html, c.genList(node))
		case *astHorizontalRuleNode:
			c.html = append(c.html, c.genHorizontalRule(node))
		case *astImageNode:
			c.html = append(c.html, c.genImage(node))
		case *astLinkNode:
			c.html = append(c.html, c.genLink(node))
		case *astCodeInlineNode:
			c.html = append(c.html, c.genCodeInline(node))
		case *astParagraphNode:
			c.html = append(c.html, c.genParagraph(node))
		default:
			return "", fmt.Errorf("unexpected node type: %T", node)
		}
	}

	return strings.Join(c.html, ""), nil
}

// genHeader emits an <hN> element for the given header node.
func (c *codegen) genHeader(node *astHeaderNode) string {
	return fmt.Sprintf("<h%d>%s</h%d>", node.size, c.genLine(node.children), node.size)
}

// genCodeBlock emits a <pre><code> block with the node's language class.
func (c *codegen) genCodeBlock(node *astCodeBlockNode) string {
	return fmt.Sprintf(`<pre><code class="%s">%s</code></pre>`, c.escapeHtml(node.lang), node.code)
}

// genQuoteBlock recursively emits a <blockquote> element for the given quote node and its children.
func (c *codegen) genQuoteBlock(node *astQuoteNode) string {
	var html []string

	html = append(html, "<blockquote>")
	for _, child := range node.children {
		switch child := child.(type) {
		case *astQuoteNode:
			html = append(html, c.genQuoteBlock(child))
		case *astQuoteItemNode:
			html = append(html, fmt.Sprintf("<p>%s</p>", c.genLine(child.children)))
		}
	}
	html = append(html, "</blockquote>")

	return strings.Join(html, "")
}

// genList emits a <ul> or <ol> element, recursing into nested list nodes.
func (c *codegen) genList(node *astListNode) string {
	var html []string

	if node.ordered {
		html = append(html, "<ol>")
	} else {
		html = append(html, "<ul>")
	}

	for _, child := range node.children {
		html = append(html, "<li>")
		for _, inner := range child.children {
			switch inner := inner.(type) {
			case *astListNode:
				html = append(html, c.genList(inner))
			case astInlineNode:
				html = append(html, c.genLine([]astInlineNode{inner}))
			}
		}
		html = append(html, "</li>")
	}

	if node.ordered {
		html = append(html, "</ol>")
	} else {
		html = append(html, "</ul>")
	}

	return strings.Join(html, "")
}

// genHorizontalRule emits a self-closing <hr> element.
func (c *codegen) genHorizontalRule(node *astHorizontalRuleNode) string {
	_ = node
	return "<hr>"
}

// genImage emits a self-closing <img> element with escaped alt and src attributes.
func (c *codegen) genImage(node *astImageNode) string {
	return fmt.Sprintf(`<img alt="%s" src="%s"/>`, c.escapeHtml(node.alt), c.escapeHtml(node.src))
}

// genLink emits an <a> element with an escaped href and text.
func (c *codegen) genLink(node *astLinkNode) string {
	return fmt.Sprintf(`<a href="%s">%s</a>`, c.escapeHtml(node.href), c.escapeHtml(node.text))
}

// genCodeInline emits a <code> element with the node's language class.
func (c *codegen) genCodeInline(node *astCodeInlineNode) string {
	return fmt.Sprintf(`<code class="%s">%s</code>`, c.escapeHtml(node.lang), node.code)
}

// genParagraph emits a <p> element containing the rendered inline children.
func (c *codegen) genParagraph(node *astParagraphNode) string {
	return fmt.Sprintf("<p>%s</p>", c.genLine(node.children))
}

// genLine renders a slice of inline nodes into a concatenated HTML string.
func (c *codegen) genLine(nodes []astInlineNode) string {
	var html []string

	for _, node := range nodes {
		switch node := node.(type) {
		case *astLinkNode:
			html = append(html, c.genLink(node))
		case *astCodeInlineNode:
			html = append(html, c.genCodeInline(node))
		case *astTextNode:
			html = append(html, c.genText(node))
		}
	}

	return strings.Join(html, "")
}

// genText emits a text node, wrapping it in <b> and/or <i> tags when bold or italic are set.
func (c *codegen) genText(node *astTextNode) string {
	var html []string

	if node.italic {
		html = append(html, "<i>")
	}
	if node.bold {
		html = append(html, "<b>")
	}

	html = append(html, c.escapeHtml(node.text))

	if node.bold {
		html = append(html, "</b>")
	}
	if node.italic {
		html = append(html, "</i>")
	}

	return strings.Join(html, "")
}

var escapeHtmlMap = map[string]string{
	"<":  "&lt;",
	">":  "&gt;",
	"&":  "&amp;",
	"\"": "&quot;",
}

// escapeHtml replaces escapable chars with their HTML entities, returning the original string unchanged if no replacements are needed.
func (c *codegen) escapeHtml(str string) string {
	var new []string

	replaced := false
	for i, ch := range str {
		replacement, ok := escapeHtmlMap[string(ch)]
		if ok {
			replaced = true

			// Copy everything up to the first replacement if this
			// is the first one we have seen.
			if len(new) == 0 {
				p := 0
				for p < i {
					new = append(new, string(str[p]))
				}
			}

			new = append(new, replacement)
		} else if len(new) > 0 {
			new = append(new, string(ch))
		}
	}

	if replaced {
		return strings.Join(new, "")
	} else {
		return str
	}
}
