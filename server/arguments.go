package server

// Argument represents an RPC method arument.
type Argument interface {
	AsString() string
	AsInt() int
	AsInt8() int8
	AsInt16() int16
	AsInt32() int32
	AsInt64() int64
	AsByte() byte
	AsBool() bool
	AsFloat32() float32
	AsFloat64() float64
	// Raw returns the argument value as a interface{}.
	Raw() interface{}
}

type argument struct {
	value interface{}
}

func (a *argument) AsString() string {
	return a.value.(string)
}

func (a *argument) AsInt() int {
	return a.value.(int)
}

func (a *argument) AsInt8() int8 {
	return a.value.(int8)
}

func (a *argument) AsInt16() int16 {
	return a.value.(int16)
}

func (a *argument) AsInt32() int32 {
	return a.value.(int32)
}

func (a *argument) AsInt64() int64 {
	return a.value.(int64)
}

func (a *argument) AsByte() byte {
	return a.value.(byte)
}

func (a *argument) AsBool() bool {
	return a.value.(bool)
}

func (a *argument) AsFloat32() float32 {
	return a.value.(float32)
}

func (a *argument) AsFloat64() float64 {
	return a.value.(float64)
}

func (a *argument) Raw() interface{} {
	return a.value
}
