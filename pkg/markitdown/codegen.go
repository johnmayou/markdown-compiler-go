package markitdown

type codegen struct{}

func newCodegen() *codegen {
	return &codegen{}
}

func (c *codegen) gen(ast astRootNode) string {
	_ = ast
	return ""
}
