package porthos

// Headers represents RPC headers (request and response).
type Headers struct {
	headers map[string]interface{}
}

// NewHeaders creates a new Headers object initializing the map.
func NewHeaders() *Headers {
	return &Headers{
		make(map[string]interface{}),
	}
}

// Set a header.
func (h *Headers) Set(key string, value interface{}) {
	h.headers[key] = value
}

// Get a header.
func (h *Headers) Get(key string) interface{} {
	return h.headers[key]
}

// Delete a header.
func (h *Headers) Delete(key string) {
	delete(h.headers, key)
}

func (h *Headers) asMap() map[string]interface{} {
	return h.headers
}
