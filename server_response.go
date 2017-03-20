package porthos

import (
	"encoding/json"
)

// Response represents a rpc response.
type Response interface {
	// JSON sets the content of the response as JSON data.
	JSON(statusCode int16, body interface{})
	// Raw sets the content of the response as an array of bytes.
	Raw(statusCode int16, contentType string, body []byte)
	// Empty leaves the content of the response as empty.
	Empty(statusCode int16)
	// GetHeaders returns the response headers.
	GetHeaders() *Headers
	// GetStatusCode returns the response status.
	GetStatusCode() int16
	GetBody() []byte
	GetContentType() string
}

type response struct {
	body        []byte
	contentType string
	statusCode  int16
	headers     *Headers
}

func newResponse() Response {
	return &response{
		headers: NewHeaders(),
	}
}

func (r *response) JSON(statusCode int16, body interface{}) {
	if body == nil {
		panic("Response body is empty")
	}

	jsonBody, err := json.Marshal(&body)

	if err != nil {
		panic(err)
	}

	r.statusCode = statusCode
	r.body = jsonBody
	r.contentType = "application/json"
}

func (r *response) Raw(statusCode int16, contentType string, body []byte) {
	if body == nil {
		panic("Response body is empty")
	}

	r.statusCode = statusCode
	r.body = body
	r.contentType = contentType
}

func (r *response) Empty(statusCode int16) {
	r.statusCode = statusCode
}

func (r *response) GetHeaders() *Headers {
	return r.headers
}

func (r *response) GetStatusCode() int16 {
	return r.statusCode
}

func (r *response) GetBody() []byte {
	return r.body
}

func (r *response) GetContentType() string {
	return r.contentType
}
