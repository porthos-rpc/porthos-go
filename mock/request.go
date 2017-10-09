package mock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/porthos-rpc/porthos-go"
)

type Request struct {
	ServiceName string
	MethodName  string
	ContentType string
	Body        []byte
	ctx         context.Context
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

func (r *Request) Form() (porthos.Form, error) {
	return porthos.NewForm(r.ContentType, r.Body)
}

func (r *Request) Bind(i interface{}) error {
	if r.ContentType != "application/json" {
		return fmt.Errorf("Invalid content type, got: %s", r.ContentType)
	}

	decoder := json.NewDecoder(bytes.NewReader(r.Body))
	decoder.UseNumber()
	return decoder.Decode(i)
}

func (r *Request) WithContext(ctx context.Context) porthos.Request {
	if ctx == nil {
		panic("nil context")
	}

	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx

	return r2
}

func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}

	return context.Background()
}

func NewRequest(service, method string, contentType string, body []byte) porthos.Request {
	return &Request{
		ServiceName: service,
		MethodName:  method,
		ContentType: contentType,
		Body:        body,
	}
}

func NewRequestFromMap(service, method string, m map[string]interface{}) porthos.Request {
	return newRequestJSON(service, method, m)
}

func NewRequestFromStruct(service, method string, i interface{}) porthos.Request {
	return newRequestJSON(service, method, i)
}

func newRequestJSON(service, method string, i interface{}) porthos.Request {
	data, err := json.Marshal(i)

	if err != nil {
		panic(err)
	}

	return NewRequest(service, method, "application/json", data)
}
