package client

import (
	"encoding/json"
	"errors"
)

// Response represents the response object of a RPC call.
type Response struct {
	StatusCode  int16
	Headers     map[string]interface{}
	Content     []byte
	ContentType string
}

// UnmarshalJSONTo outputs the response content to the argument pointer.
func (r *Response) UnmarshalJSONTo(v interface{}) error {
	if r.ContentType != "application/json" {
		return errors.New("The content type of the response is not 'application/json'")
	}

	err := json.Unmarshal(r.Content, v)
	return err
}

// UnmarshalJSON outputs the response content to the argument pointer.
func (r *Response) UnmarshalJSON() (map[string]interface{}, error) {
	if r.ContentType != "application/json" {
		return nil, errors.New("The content type of the response is not 'application/json'")
	}

	var v map[string]interface{}
	err := json.Unmarshal(r.Content, &v)

	return v, err
}
