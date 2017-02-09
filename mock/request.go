package mock

import (
	"github.com/porthos-rpc/porthos-go/server"
)

func NewRequest(service, method string, args ...interface{}) server.Request {
	return &Request{
		ServiceName: service,
		MethodName:  method,
		Args:        args,
	}
}

type Request struct {
	ServiceName string
	MethodName  string
	Args        []interface{}
}

func (r *Request) GetServiceName() string {
	return r.ServiceName
}

func (r *Request) GetMethodName() string {
	return r.MethodName
}

func (r *Request) GetRawArgs() []interface{} {
	return r.Args
}

func (r *Request) GetArg(index int) server.Argument {
	return server.NewArgument(r.Args[index])
}
