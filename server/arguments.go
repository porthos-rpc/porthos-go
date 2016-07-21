package server

// Argument represents an RPC method arument.
type Argument struct {
    value interface{}
}

func (a *Argument) AsString() string {
    return a.value.(string)
}

func (a *Argument) AsInt() int {
    return a.value.(int)
}

func (a *Argument) AsInt8() int8 {
    return a.value.(int8)
}

func (a *Argument) AsInt16() int16 {
    return a.value.(int16)
}

func (a *Argument) AsInt32() int32 {
    return a.value.(int32)
}

func (a *Argument) AsInt64() int64 {
    return a.value.(int64)
}

func (a *Argument) AsByte() byte {
    return a.value.(byte)
}

func (a *Argument) AsBool() bool {
    return a.value.(bool)
}

func (a *Argument) AsFloat32() float32 {
    return a.value.(float32)
}

func (a *Argument) AsFloat64() float64 {
    return a.value.(float64)
}

// Raw returns the argument value as a interface{}
func (a *Argument) Raw() interface{} {
    return a.value
}
