package server

// Request represents a rpc request.
type Request interface {
	// GetServiceName returns the service name.
	GetServiceName() string
	// GetMethodName returns the method name.
	GetMethodName() string
	// GetBody returns the request body.
	GetBody() []byte
	// IndexForm returns a index-based form.
	IndexForm() (IndexForm, error)
	// MapForm returns a map-based form.
	MapForm() (MapForm, error)
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

func (r *request) IndexForm() (IndexForm, error) {
	return NewIndexForm(r.contentType, r.body)
}

func (r *request) MapForm() (MapForm, error) {
	return NewMapForm(r.contentType, r.body)
}
