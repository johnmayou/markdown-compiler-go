package markitdown

type astNode interface {
	isNode()
}

type astRootNode struct {
	children []astNode
}

func (n *astRootNode) isNode() {}

type parser struct {
	tks []token
}

func newParser(tks []token) *parser {
	return &parser{tks: tks}
}

func (p *parser) parse() (astRootNode, error) {
	_ = p.tks
	return astRootNode{children: make([]astNode, 0)}, nil
}
