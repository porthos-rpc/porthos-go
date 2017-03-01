package mock

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/porthos-rpc/porthos-go/server"
)

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

func (r *Request) Form() (server.Form, error) {
	return server.NewForm(r.ContentType, r.Body)
}

func (r *Request) Bind(i interface{}) error {
	if r.ContentType != "application/json" {
		return fmt.Errorf("Invalid content type, got: %s", r.ContentType)
	}

	decoder := json.NewDecoder(bytes.NewReader(r.Body))
	decoder.UseNumber()
	return decoder.Decode(i)
}
