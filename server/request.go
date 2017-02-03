package server

// Request represents a rpc request.
type Request interface {
	// GetServiceName returns the service name.
	GetServiceName() string
	// GetMethodName returns the method name.
	GetMethodName() string
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

func (r *request) GetServiceName() string {
	return r.serviceName
}

func (r *request) GetMethodName() string {
	return r.methodName
}

func (r *request) GetRawArgs() []interface{} {
	return r.args
}

func (r *request) GetArg(index int) Argument {
	return &argument{r.args[index]}
}
