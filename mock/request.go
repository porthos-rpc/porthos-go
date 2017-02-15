package mock

import "github.com/porthos-rpc/porthos-go/server"

func NewRequest(service, method string, contentType string, body []byte) *Request {
	return &Request{
		ServiceName: service,
		MethodName:  method,
		ContentType: contentType,
		Body:        body,
	}
}

type Request struct {
	ServiceName string
	MethodName  string
	ContentType string
	Body        []byte
}

func (r *Request) GetServiceName() string {
	return r.ServiceName
}

func (r *Request) GetMethodName() string {
	return r.MethodName
}

func (r *Request) GetBody() []byte {
	return r.Body
}

func (r *Request) IndexForm() (server.IndexForm, error) {
	return server.NewIndexForm(r.ContentType, r.Body)
}

func (r *Request) MapForm() (server.MapForm, error) {
	return server.NewMapForm(r.ContentType, r.Body)
}
