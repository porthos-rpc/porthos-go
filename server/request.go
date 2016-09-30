package server

// Request represents a rpc request.
type Request struct {
	ServiceName string
	MethodName  string
	args        []interface{}
	messageBody []byte
}

// GetArg returns an argument giving the index.
func (r *Request) GetArg(index int) *Argument {
	return &Argument{r.args[index]}
}
