package porthos

import "reflect"

// Spec to a remote procedure.
type Spec struct {
	ContentType string      `json:"contentType"`
	Body        BodySpecMap `json:"body"`
}

// BodySpecMap represents a body spec.
type BodySpecMap map[string]interface{}

// BodySpecFromStruct creates a body spec from a struct value.
// You just have to pass an "instance" of your struct.
func BodySpecFromStruct(structValue interface{}) BodySpecMap {
	return BodySpecFromStructType(reflect.ValueOf(structValue).Type())
}

// BodySpecFromStructType creates a body spec from a struct type (reflect).
func BodySpecFromStructType(s reflect.Type) BodySpecMap {
	spec := BodySpecMap{}

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		jsonField := f.Tag.Get("json")

		if f.Type.Kind() == reflect.Struct {
			spec[jsonField] = BodySpecFromStructType(f.Type)
		} else {
			spec[jsonField] = f.Type.Name()
		}
	}

	return spec
}
