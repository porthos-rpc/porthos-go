package mock

import (
	"encoding/json"

	"github.com/porthos-rpc/porthos-go"
)

func NewResponse() porthos.Response {
	return &Response{
		Headers: porthos.NewHeaders(),
	}
}

type Response struct {
	Body        []byte
	ContentType string
	StatusCode  int32
	Headers     *porthos.Headers
}

func (r *Response) JSON(statusCode int32, Body interface{}) {
	jsonBody, err := json.Marshal(&Body)

	if err != nil {
		panic(err)
	}

	r.StatusCode = statusCode
	r.Body = jsonBody
	r.ContentType = "application/json"
}

func (r *Response) Raw(statusCode int32, contentType string, body []byte) {
	r.StatusCode = statusCode
	r.Body = body
	r.ContentType = contentType
}

func (r *Response) Empty(statusCode int32) {
	r.StatusCode = statusCode
}

func (r *Response) GetHeaders() *porthos.Headers {
	return r.Headers
}

func (r *Response) GetStatusCode() int32 {
	return r.StatusCode
}

func (r *Response) GetBody() []byte {
	return r.Body
}

func (r *Response) GetContentType() string {
	return r.ContentType
}
