package mock

type Argument struct {
	Value interface{}
}

func (a *Argument) AsString() string {
	return a.Value.(string)
}

func (a *Argument) AsInt() int {
	return a.Value.(int)
}

func (a *Argument) AsInt8() int8 {
	return a.Value.(int8)
}

func (a *Argument) AsInt16() int16 {
	return a.Value.(int16)
}

func (a *Argument) AsInt32() int32 {
	return a.Value.(int32)
}

func (a *Argument) AsInt64() int64 {
	return a.Value.(int64)
}

func (a *Argument) AsByte() byte {
	return a.Value.(byte)
}

func (a *Argument) AsBool() bool {
	return a.Value.(bool)
}

func (a *Argument) AsFloat32() float32 {
	return a.Value.(float32)
}

func (a *Argument) AsFloat64() float64 {
	return a.Value.(float64)
}

func (a *Argument) Raw() interface{} {
	return a.Value
}
