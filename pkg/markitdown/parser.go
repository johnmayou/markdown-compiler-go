package markitdown

type astNode interface {
	isNode()
}

type astRootNode struct {
	children []astNode
}

func (n *astRootNode) isNode() {}

type parser struct{}

func newParser() *parser {
	return &parser{}
}

func (p *parser) parse(tks []token) astRootNode {
	_ = tks
	return astRootNode{
		children: make([]astNode, 0),
	}
}
