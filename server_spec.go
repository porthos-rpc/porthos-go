package porthos

import "reflect"

// Spec to a remote procedure.
type Spec struct {
	Description string      `json:"description"`
	Request     ContentSpec `json:"request"`
	Response    ContentSpec `json:"response"`
}

// ContentSpec to a remote procedure.
type ContentSpec struct {
	ContentType string      `json:"contentType"`
	Body        interface{} `json:"body"`
}

// FieldSpec represents a spec of a body field.
type FieldSpec struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Body        BodySpecMap `json:"body,omitempty"`
}

// BodySpecMap represents a body spec.
type BodySpecMap map[string]FieldSpec

// BodySpecFromStruct creates a body spec from a struct value.
// You just have to pass an "instance" of your struct.
func BodySpecFromStruct(structValue interface{}) BodySpecMap {
	return bodySpecFromStructType(reflect.ValueOf(structValue).Type())
}

// BodySpecFromArray creates a body spec from a array value.
// You just have to pass an "instance" of your array.
func BodySpecFromArray(structOfTheList interface{}) []FieldSpec {
	spec := make([]FieldSpec, 1)

	tp := reflect.ValueOf(structOfTheList).Type()

	if tp.Kind() == reflect.Struct {
		child := bodySpecFromStructType(tp)
		spec[0] = FieldSpec{Type: tp.Name(), Body: child}
	} else {
		spec[0] = FieldSpec{Type: tp.Name()}
	}

	return spec
}

func bodySpecFromStructType(s reflect.Type) BodySpecMap {
	spec := BodySpecMap{}

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		jsonField := f.Tag.Get("json")
		description := f.Tag.Get("description")

		if f.Type.Kind() == reflect.Struct {
			child := bodySpecFromStructType(f.Type)
			spec[jsonField] = FieldSpec{Type: f.Type.Name(), Description: description, Body: child}
		} else {
			spec[jsonField] = FieldSpec{Type: f.Type.Name(), Description: description}
		}
	}

	return spec
}
