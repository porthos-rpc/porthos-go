package server

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// IndexForm represents a request form where data are retrieved through indexes.
type IndexForm interface {
	// GetArg returns an argument giving the index.
	GetArg(index int) Argument
}

// MapForm represents a request form where data are retrieved through field names.
type MapForm interface {
	// GetArg returns an argument giving the index.
	GetArg(field string) Argument
}

type indexForm struct {
	args []interface{}
}

type mapForm struct {
	fields map[string]interface{}
}

// NewIndexForm creates a new form to retrieve values from its index.
func NewIndexForm(contentType string, body []byte) (IndexForm, error) {
	if contentType != "application/porthos-args" {
		return nil, fmt.Errorf("Invalid contentType, got: %s", contentType)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	args := new([]interface{})
	err := decoder.Decode(args)

	if err != nil {
		return nil, err
	}

	return &indexForm{*args}, nil
}

// NewMapForm creates a new form to retrieve values from field names.
func NewMapForm(contentType string, body []byte) (MapForm, error) {
	if contentType != "application/porthos-map" {
		return nil, fmt.Errorf("Invalid contentType, got: %s", contentType)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	fields := new(map[string]interface{})
	err := decoder.Decode(fields)

	if err != nil {
		return nil, err
	}

	return &mapForm{*fields}, nil
}

func (r *indexForm) GetArg(index int) Argument {
	return &argument{r.args[index]}
}

func (r *mapForm) GetArg(field string) Argument {
	return &argument{r.fields[field]}
}
