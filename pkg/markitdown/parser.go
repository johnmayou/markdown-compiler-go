package markitdown

import (
	"fmt"
)

type astNode interface {
	isNode()
}

type astInlineNode interface {
	astNode
	isInlineNode()
}

type astQuoteChild interface {
	astNode
	isQuoteChild()
}

type astRootNode struct {
	children []astNode
}

type astHeaderNode struct {
	size     int
	children []astInlineNode
}

type astCodeBlockNode struct {
	lang string
	code string
}

type astCodeInlineNode struct {
	lang string
	code string
}

type astQuoteNode struct {
	children []astQuoteChild
}

type astQuoteItemNode struct {
	children []astInlineNode
}

type astParagraphNode struct {
	children []astInlineNode
}

type astTextNode struct {
	text   string
	bold   bool
	italic bool
}

type astHorizontalRuleNode struct{}

type astImageNode struct {
	alt string
	src string
}

type astLinkNode struct {
	text string
	href string
}

type astListNode struct {
	ordered  bool
	children []astListItemNode
}

type astListItemNode struct {
	children []astNode
}

func (n *astRootNode) isNode()           {}
func (n *astHeaderNode) isNode()         {}
func (n *astCodeBlockNode) isNode()      {}
func (n *astCodeInlineNode) isNode()     {}
func (n *astQuoteNode) isNode()          {}
func (n *astQuoteItemNode) isNode()      {}
func (n *astParagraphNode) isNode()      {}
func (n *astTextNode) isNode()           {}
func (n *astHorizontalRuleNode) isNode() {}
func (n *astImageNode) isNode()          {}
func (n *astLinkNode) isNode()           {}
func (n *astListNode) isNode()           {}
func (n *astListItemNode) isNode()       {}

func (n *astCodeInlineNode) isInlineNode() {}
func (n *astTextNode) isInlineNode()       {}
func (n *astImageNode) isInlineNode()      {}
func (n *astLinkNode) isInlineNode()       {}

func (n *astQuoteNode) isQuoteChild()     {}
func (n *astQuoteItemNode) isQuoteChild() {}

type parser struct {
	tks      []token
	tksStart int
	root     *astRootNode
}

// newParser returns a parser initialized with the given token slice.
func newParser(tks []token) *parser {
	return &parser{tks: tks, tksStart: 0, root: &astRootNode{}}
}

// parse processes all tokens into an AST rooted at astRootNode.
func (p *parser) parse() (*astRootNode, error) {
	for p.tksStart < len(p.tks) {
		if peek[*headerToken](p, 1) {
			node, err := p.parseHeader()
			if err != nil {
				return p.root, fmt.Errorf("parsing header: %w", err)
			}
			p.root.children = append(p.root.children, node)
		} else if peek[*codeBlockToken](p, 1) {
			node, err := p.parseCodeBlock()
			if err != nil {
				return p.root, fmt.Errorf("parsing code block: %w", err)
			}
			p.root.children = append(p.root.children, node)
		} else if peek[*blockQuoteToken](p, 1) {
			node, err := p.parseBlockQuote()
			if err != nil {
				return p.root, fmt.Errorf("parsing quote block: %w", err)
			}
			p.root.children = append(p.root.children, node)
		} else if peek[*horizontalRuleToken](p, 1) {
			node, err := p.parseHorizontalRule()
			if err != nil {
				return p.root, fmt.Errorf("parsing horizontal rule: %w", err)
			}
			p.root.children = append(p.root.children, node)
		} else if peek[*listItemToken](p, 1) {
			node, err := p.parseList()
			if err != nil {
				return p.root, fmt.Errorf("parsing list: %w", err)
			}
			p.root.children = append(p.root.children, node)
		} else if peek[*imageToken](p, 1) {
			node, err := p.parseImage()
			if err != nil {
				return p.root, fmt.Errorf("parsing image: %w", err)
			}
			p.root.children = append(p.root.children, node)
		} else if peek[*textToken](p, 1) || peek[*codeInlineToken](p, 1) || peek[*linkToken](p, 1) {
			node, err := p.parseParagraph()
			if err != nil {
				return p.root, fmt.Errorf("parsing paragraph: %w", err)
			}
			p.root.children = append(p.root.children, node)
		} else if peek[*newLineToken](p, 1) {
			if _, err := consume[*newLineToken](p); err != nil {
				return p.root, fmt.Errorf("consume new line: %w", err)
			}
		} else {
			return p.root, fmt.Errorf("unsupported token: %T", p.tks[p.tksStart])
		}
	}

	return p.root, nil
}

// parseHeader consumes a headerToken and its inline children, returning an astHeaderNode.
func (p *parser) parseHeader() (*astHeaderNode, error) {
	t, err := consume[*headerToken](p)
	if err != nil {
		return nil, fmt.Errorf("consuming header: %w", err)
	}
	children, err := p.parseInline()
	if err != nil {
		return nil, fmt.Errorf("parsing header children: %w", err)
	}
	return &astHeaderNode{size: t.size, children: children}, nil
}

// parseCodeBlock consumes a codeBlockToken and its trailing newline, returning an astCodeBlockNode.
func (p *parser) parseCodeBlock() (*astCodeBlockNode, error) {
	t, err := consume[*codeBlockToken](p)
	if err != nil {
		return nil, fmt.Errorf("consuming code block: %w", err)
	}
	if _, err = consume[*newLineToken](p); err != nil {
		return nil, fmt.Errorf("consuming code block new line: %w", err)
	}
	return &astCodeBlockNode{lang: t.lang, code: t.code}, nil
}

// parseBlockQuote consumes a sequence of blockQuoteTokens and builds a nested astQuoteNode tree keyed by indent level.
func (p *parser) parseBlockQuote() (*astQuoteNode, error) {
	quote, err := consume[*blockQuoteToken](p)
	if err != nil {
		return nil, fmt.Errorf("consuming root quote: %w", err)
	}
	quoteRootItem, err := p.parseQuoteItem()
	if err != nil {
		return nil, fmt.Errorf("consuming root quote children: %w", err)
	}
	quoteRoot := &astQuoteNode{children: []astQuoteChild{quoteRootItem}}

	nodeIndentMap := make(map[int]*astQuoteNode)
	nodeIndentMap[quote.indent] = quoteRoot

	for peek[*blockQuoteToken](p, 1) {
		t, err := consume[*blockQuoteToken](p)
		if err != nil {
			return nil, fmt.Errorf("invariant violation: peek confirmed token but consume failed: %w", err)
		}

		if peek[*newLineToken](p, 1) {
			if _, err := consume[*newLineToken](p); err != nil {
				return nil, fmt.Errorf("invariant violation: peek confirmed token but consume failed: %w", err)
			}
			continue
		}

		node, ok := nodeIndentMap[t.indent]
		if ok {
			item, err := p.parseQuoteItem()
			if err != nil {
				return nil, fmt.Errorf("parsing quote item: %w", err)
			}
			node.children = append(node.children, item)
		} else {
			item, err := p.parseQuoteItem()
			if err != nil {
				return nil, fmt.Errorf("parsing quote item: %w", err)
			}
			node := &astQuoteNode{children: []astQuoteChild{item}}
			nodeIndentMap[t.indent] = node

			// Add new node to parent quote children.
			parent, ok := nodeIndentMap[t.indent-1]
			if !ok {
				parent = quoteRoot
			}
			parent.children = append(parent.children, node)
		}
	}

	return quoteRoot, nil
}

// parseQuoteItem parses inline content following a blockQuoteToken into an astQuoteItemNode.
func (p *parser) parseQuoteItem() (*astQuoteItemNode, error) {
	children, err := p.parseInlineBlockQuote()
	if err != nil {
		return nil, fmt.Errorf("parsing block quote item children: %w", err)
	}
	return &astQuoteItemNode{children: children}, nil
}

type listStackItem struct {
	node   *astListNode
	indent int
}

// parseList consumes a run of listItemTokens and builds a nested astListNode tree using a stack to track indent levels.
func (p *parser) parseList() (*astListNode, error) {
	rootItem, err := consume[*listItemToken](p)
	if err != nil {
		return nil, fmt.Errorf("consume root list item: %w", err)
	}
	rootChildren, err := p.parseInline()
	if err != nil {
		return nil, fmt.Errorf("parsing root list item children: %w", err)
	}
	root := &astListNode{ordered: rootItem.ordered, children: []astListItemNode{{children: inlineToNodes(rootChildren)}}}

	listStack := make([]listStackItem, 0)
	listStack = append(listStack, listStackItem{node: root, indent: 0})

	for peek[*listItemToken](p, 1) {
		curr, err := consume[*listItemToken](p)
		if err != nil {
			return nil, fmt.Errorf("invariant violation: peek confirmed token but consume failed: %w", err)
		}
		currIndent := min(listStack[len(listStack)-1].indent+1, curr.indent)
		lastIndent := listStack[len(listStack)-1].indent

		if currIndent > lastIndent { // Deeper indent.
			// Create new node.
			children, err := p.parseInline()
			if err != nil {
				return nil, fmt.Errorf("parsing list item children: %w", err)
			}
			node := &astListNode{ordered: curr.ordered, children: []astListItemNode{{inlineToNodes(children)}}}

			// Append to last child of the top of stack node.
			top := listStack[len(listStack)-1].node
			topLastChild := top.children[len(top.children)-1]
			topLastChild.children = append(topLastChild.children, node)

			// Add new node to stack.
			listStack = append(listStack, listStackItem{node: node, indent: currIndent})
		} else if currIndent < lastIndent { // Lost indentation.
			// Pop from stack until we find the current level.
			for listStack[len(listStack)-1].indent > currIndent {
				listStack = listStack[:len(listStack)-1]
			}

			children, err := p.parseInline()
			if err != nil {
				return nil, fmt.Errorf("parsing list item children: %w", err)
			}
			last := listStack[len(listStack)-1]
			last.node.children = append(last.node.children, astListItemNode{children: inlineToNodes(children)})
		} else { // Same indentation.
			children, err := p.parseInline()
			if err != nil {
				return nil, fmt.Errorf("parsing list item children: %w", err)
			}
			last := listStack[len(listStack)-1]
			last.node.children = append(last.node.children, astListItemNode{children: inlineToNodes(children)})
		}
	}

	return root, nil
}

// parseHorizontalRule consumes a horizontalRuleToken and its trailing newline, returning an astHorizontalRuleNode.
func (p *parser) parseHorizontalRule() (*astHorizontalRuleNode, error) {
	_, err := consume[*horizontalRuleToken](p)
	if err != nil {
		return nil, fmt.Errorf("consuming horizontal rule: %w", err)
	}
	_, err = consume[*newLineToken](p)
	if err != nil {
		return nil, fmt.Errorf("consuming new line: %w", err)
	}
	return &astHorizontalRuleNode{}, nil
}

// parseImage consumes an imageToken and its trailing newline, returning an astImageNode.
func (p *parser) parseImage() (*astImageNode, error) {
	t, err := consume[*imageToken](p)
	if err != nil {
		return nil, fmt.Errorf("consuming image: %w", err)
	}
	_, err = consume[*newLineToken](p)
	if err != nil {
		return nil, fmt.Errorf("consuming new line: %w", err)
	}
	return &astImageNode{alt: t.alt, src: t.src}, nil
}

// parseParagraph parses a run of inline tokens into an astParagraphNode.
func (p *parser) parseParagraph() (*astParagraphNode, error) {
	children, err := p.parseInline()
	if err != nil {
		return nil, fmt.Errorf("parsing paragraph children: %w", err)
	}
	return &astParagraphNode{children: children}, nil
}

// parseInline collects inline tokens across soft line breaks until a blank line or block-level token is reached.
func (p *parser) parseInline() ([]astInlineNode, error) {
	var nodes []astInlineNode

	onInlineLine := func() bool {
		// Inline tokens on current line.
		if peekInline(p, 1) {
			return true
		}

		// A newline followed by inline tokens is a soft line break,
		// considered part of the same component.
		return peek[*newLineToken](p, 1) && peekInline(p, 2)
	}

	for onInlineLine() {
		if peek[*newLineToken](p, 1) {
			if _, err := consume[*newLineToken](p); err != nil {
				return nil, fmt.Errorf("consuming new line: %w", err)
			}
			nodes = append(nodes, &astTextNode{text: " ", bold: false, italic: false})
		}

		node, err := p.parseInlineSingle()
		if err != nil {
			return nil, fmt.Errorf("parsing inline value: %w", err)
		}
		nodes = append(nodes, node)
	}
	if _, err := consume[*newLineToken](p); err != nil {
		return nil, fmt.Errorf("consuming new line: %w", err)
	}

	return nodes, nil
}

// parseInlineBlockQuote collects inline tokens within a block quote, treating a newline+blockQuoteToken pair as a soft line break.
func (p *parser) parseInlineBlockQuote() ([]astInlineNode, error) {
	var nodes []astInlineNode

	onInlineLine := func() bool {
		// Inline tokens on current line.
		if peekInline(p, 1) {
			return true
		}

		// A newline followed by another block quote and inline tokens
		// is a soft line break, considered part of the same component.
		return peek[*newLineToken](p, 1) && peek[*blockQuoteToken](p, 2) && peekInline(p, 3)
	}

	for onInlineLine() {
		if peek[*newLineToken](p, 1) {
			if _, err := consume[*newLineToken](p); err != nil {
				return nil, fmt.Errorf("consuming new line: %w", err)
			}
			if _, err := consume[*blockQuoteToken](p); err != nil {
				return nil, fmt.Errorf("consuming quote block: %w", err)
			}
			nodes = append(nodes, &astTextNode{text: " ", bold: false, italic: false})
		}

		node, err := p.parseInlineSingle()
		if err != nil {
			return nil, fmt.Errorf("parsing inline value: %w", err)
		}
		nodes = append(nodes, node)
	}
	if _, err := consume[*newLineToken](p); err != nil {
		return nil, fmt.Errorf("consuming new line: %w", err)
	}

	return nodes, nil
}

// parseInlineSingle consumes exactly one inline token (text, inline code, or link) and returns the corresponding AST node.
func (p *parser) parseInlineSingle() (astInlineNode, error) {
	if peek[*textToken](p, 1) {
		t, err := consume[*textToken](p)
		if err != nil {
			return nil, fmt.Errorf("invariant violation: peek confirmed token but consume failed: %w", err)
		}
		return &astTextNode{text: t.text, bold: t.bold, italic: t.italic}, nil
	} else if peek[*codeInlineToken](p, 1) {
		t, err := consume[*codeInlineToken](p)
		if err != nil {
			return nil, fmt.Errorf("invariant violation: peek confirmed token but consume failed: %w", err)
		}
		return &astCodeInlineNode{lang: t.lang, code: t.code}, nil
	} else if peek[*linkToken](p, 1) {
		t, err := consume[*linkToken](p)
		if err != nil {
			return nil, fmt.Errorf("invariant violation: peek confirmed token but consume failed: %w", err)
		}
		return &astLinkNode{text: t.text, href: t.href}, nil
	} else {
		return nil, fmt.Errorf("expected inline token type but got: %T", p.tks[p.tksStart])
	}
}

// peekInline peeks for tokens that are considered "inline" for paragraphs, block quote items, etc.
func peekInline(p *parser, depth int) bool {
	return peek[*textToken](p, depth) || peek[*codeInlineToken](p, depth) || peek[*linkToken](p, depth)
}

// peek reports whether the token at the given lookahead depth is of type T without consuming it.
func peek[T token](p *parser, depth int) bool {
	index := p.tksStart + depth - 1
	if index >= len(p.tks) {
		return false
	}
	_, ok := p.tks[index].(T)
	return ok
}

// consume returns the next token as type T and advances the cursor, returning an error if the queue is empty or the type does not match.
func consume[T token](p *parser) (T, error) {
	var zero T
	if p.tksStart >= len(p.tks) {
		return zero, fmt.Errorf("no tokens to consume")
	}

	t, ok := p.tks[p.tksStart].(T)
	if !ok {
		return zero, fmt.Errorf("expected %T token type but got %T", zero, p.tks[p.tksStart])
	}

	p.tksStart++
	return t, nil
}

func inlineToNodes(inline []astInlineNode) []astNode {
	nodes := make([]astNode, len(inline))
	for i, n := range inline {
		nodes[i] = n
	}
	return nodes
}
