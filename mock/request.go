package mock

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/porthos-rpc/porthos-go"
)

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

func NewRequest(service, method string, contentType string, body []byte) *Request {
	return &Request{
		ServiceName: service,
		MethodName:  method,
		ContentType: contentType,
		Body:        body,
	}
}

func NewRequestFromMap(service, method string, m map[string]interface{}) *Request {
	return newRequestJSON(service, method, m)
}

func NewRequestFromStruct(service, method string, i interface{}) *Request {
	return newRequestJSON(service, method, i)
}

func newRequestJSON(service, method string, i interface{}) *Request {
	data, err := json.Marshal(i)

	if err != nil {
		panic(err)
	}

	return NewRequest(service, method, "application/json", data)
}
