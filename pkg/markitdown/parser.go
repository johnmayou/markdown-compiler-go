package markitdown

type nodeKind int

const (
	nodeRoot nodeKind = iota
)

type astNode interface {
	kind() nodeKind
}

type astRootNode struct {
	children []astNode
}

func (n *astRootNode) kind() nodeKind { return nodeRoot }

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
