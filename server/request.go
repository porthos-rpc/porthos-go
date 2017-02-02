package server

// Request represents a rpc request.
type Request interface {
	// ServiceName returns the service name.
	ServiceName() string
	// MethodName returns the method name.
	MethodName() string
	// GetRawArgs returns all arguments.
	GetRawArgs() []interface{}
	// GetArg returns an argument giving the index.
	GetArg(index int) Argument
}

type request struct {
	serviceName string
	methodName  string
	args        []interface{}
}

func (r *request) ServiceName() string {
	return r.serviceName
}

func (r *request) MethodName() string {
	return r.methodName
}

func (r *request) GetRawArgs() []interface{} {
	return r.args
}

func (r *request) GetArg(index int) Argument {
	return &argument{r.args[index]}
}
