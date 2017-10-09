package porthos

import (
	"bytes"
	"context"
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
	// WithContext returns a shallow copy of Event with its context changed to context.
	// The provided context must be non-nil.
	WithContext(context.Context) Request
	// The returned context is always non-nil; it defaults to the background context.
	// To change the context, use WithContext.
	Context() context.Context
}

type request struct {
	serviceName string
	methodName  string
	contentType string
	body        []byte
	ctx         context.Context
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

func (r *request) WithContext(ctx context.Context) Request {
	if ctx == nil {
		panic("nil context")
	}

	r2 := new(request)
	*r2 = *r
	r2.ctx = ctx

	return r2
}

func (r *request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}

	return context.Background()
}
