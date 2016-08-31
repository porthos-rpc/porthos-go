package server

// Request represents a rpc request.
type Request struct {
	MethodName string
	args       []interface{}
}

// GetArg returns an argument giving the index.
func (r *Request) GetArg(index int) *Argument {
	return &Argument{r.args[index]}
}
