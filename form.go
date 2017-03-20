package porthos

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Form represents a request form where data are retrieved through indexes.
type Form interface {
	// GetArg returns an argument giving the index.
	GetArg(index int) Argument
}

type form struct {
	args []interface{}
}

// NewForm creates a new form to retrieve values from its index.
func NewForm(contentType string, body []byte) (Form, error) {
	if contentType != "application/json" {
		return nil, fmt.Errorf("Invalid content type, got: %s", contentType)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	args := new([]interface{})
	err := decoder.Decode(args)

	if err != nil {
		return nil, err
	}

	return &form{*args}, nil
}

func (r *form) GetArg(index int) Argument {
	return &argument{r.args[index]}
}
