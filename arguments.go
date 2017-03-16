package porthos

import (
	"encoding/json"
	"errors"
)

// Argument represents an RPC method arument.
type Argument interface {
	AsString() (string, error)
	AsInt() (int, error)
	AsInt8() (int8, error)
	AsInt16() (int16, error)
	AsInt32() (int32, error)
	AsInt64() (int64, error)
	AsByte() (byte, error)
	AsBool() (bool, error)
	AsFloat32() (float32, error)
	AsFloat64() (float64, error)
	// Raw returns the argument value as a interface{}.
	Raw() interface{}
}

type argument struct {
	value interface{}
}

var (
	// ErrTypeCast returned when a type cast fails.
	ErrTypeCast = errors.New("Error reading string argument")
)

// NewArgument creates a new Argument.
func NewArgument(value interface{}) Argument {
	return &argument{value}
}

func (a *argument) AsString() (string, error) {
	s, valid := a.value.(string)

	if valid {
		return s, nil
	}

	return s, ErrTypeCast
}

func (a *argument) AsInt() (int, error) {
	i, err := a.value.(json.Number).Int64()
	return int(i), err
}

func (a *argument) AsInt8() (int8, error) {
	i, err := a.value.(json.Number).Int64()
	return int8(i), err
}

func (a *argument) AsInt16() (int16, error) {
	i, err := a.value.(json.Number).Int64()
	return int16(i), err
}

func (a *argument) AsInt32() (int32, error) {
	i, err := a.value.(json.Number).Int64()
	return int32(i), err
}

func (a *argument) AsInt64() (int64, error) {
	i, err := a.value.(json.Number).Int64()
	return i, err
}

func (a *argument) AsByte() (byte, error) {
	b, valid := a.value.(byte)

	if valid {
		return b, nil
	}

	return b, ErrTypeCast
}

func (a *argument) AsBool() (bool, error) {
	b, valid := a.value.(bool)

	if valid {
		return b, nil
	}

	return b, ErrTypeCast
}

func (a *argument) AsFloat32() (float32, error) {
	f, err := a.value.(json.Number).Float64()
	return float32(f), err
}

func (a *argument) AsFloat64() (float64, error) {
	f, err := a.value.(json.Number).Float64()
	return f, err
}

func (a *argument) Raw() interface{} {
	return a.value
}
