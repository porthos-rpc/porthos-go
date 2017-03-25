package porthos

import "reflect"

// Spec to a remote procedure.
type Spec struct {
	ContentType string
	Body        BodySpecMap
}

// BodySpecMap represents a body spec.
type BodySpecMap map[string]interface{}

// BodySpecFromStruct creates a body spec from a struct value.
// You just have to pass an "instance" of your struct.
func BodySpecFromStruct(structValue interface{}) BodySpecMap {
	spec := BodySpecMap{}

	s := reflect.ValueOf(structValue)
	t := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		jsonField := t.Field(i).Tag.Get("json")

		if jsonField != "" {
			spec[jsonField] = f.Type()
		}
	}

	return spec
}
