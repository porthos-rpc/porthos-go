package porthos

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Request represents a rpc request.
type Request interface {
	// GetServiceName returns the service name.
	GetServiceName() string
	// GetMethodName returns the method name.
	GetMethodName() string
	// GetBody returns the request body.
	GetBody() []byte
	// Form returns a index-based form.
	Form() (Form, error)
	// Bind binds the body to an interface.
	Bind(i interface{}) error
}

type request struct {
	serviceName string
	methodName  string
	contentType string
	body        []byte
}

func (r *request) GetServiceName() string {
	return r.serviceName
}

func (r *request) GetMethodName() string {
	return r.methodName
}

func (r *request) GetBody() []byte {
	return r.body
}

func (r *request) Form() (Form, error) {
	return NewForm(r.contentType, r.body)
}

func (r *request) Bind(i interface{}) error {
	if r.contentType != "application/json" {
		return fmt.Errorf("Invalid content type, got: %s", r.contentType)
	}

	decoder := json.NewDecoder(bytes.NewReader(r.body))
	decoder.UseNumber()
	return decoder.Decode(i)
}
