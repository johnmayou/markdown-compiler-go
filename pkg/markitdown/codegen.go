package markitdown

type codegen struct {
	ast *astRootNode
}

func newCodegen(ast *astRootNode) *codegen {
	return &codegen{ast: ast}
}

func (c *codegen) gen() (string, error) {
	_ = c.ast
	return "", nil
}
