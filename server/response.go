package server

import (
	"encoding/json"
)

// Response represents a rpc response.
type Response interface {
	// JSON sets the content of the response as JSON data.
	JSON(statusCode int16, content interface{})
	// Raw sets the content of the response as an array of bytes.
	Raw(statusCode int16, contentType string, content []byte)
	// Empty leaves the content of the response as empty.
	Empty(statusCode int16)
	// Headers returns the response headers.
	Headers() *Headers
	// StatusCode returns the response status.
	StatusCode() int16
	Body() []byte
	ContentType() string
}

type response struct {
	content     []byte
	contentType string
	statusCode  int16
	headers     *Headers
}

func newResponse() Response {
	return &response{
		headers: NewHeaders(),
	}
}

func (r *response) JSON(statusCode int16, content interface{}) {
	if content == nil {
		panic("Response content is empty")
	}

	jsonContent, err := json.Marshal(&content)

	if err != nil {
		panic(err)
	}

	r.statusCode = statusCode
	r.content = jsonContent
	r.contentType = "application/json"
}

func (r *response) Raw(statusCode int16, contentType string, content []byte) {
	if content == nil {
		panic("Response content is empty")
	}

	r.statusCode = statusCode
	r.content = content
	r.contentType = contentType
}

func (r *response) Empty(statusCode int16) {
	r.statusCode = statusCode
}

func (r *response) Headers() *Headers {
	return r.headers
}

func (r *response) StatusCode() int16 {
	return r.statusCode
}

func (r *response) Body() []byte {
	return r.content
}

func (r *response) ContentType() string {
	return r.contentType
}
