package porthos

import (
	"reflect"
	"testing"
)

func TestBodySpecFromStruct(t *testing.T) {
	type X struct {
		FloatArg float32 `json:"float_arg"`
	}

	type Y struct {
		Original  int    `json:"original_value"`
		Sum       int32  `json:"value_plus_one"`
		StrArg    string `json:"str_arg"`
		BoolArg   bool   `json:"bool_arg"`
		StructArg X      `json:"struct_arg"`
	}

	bodySpec := BodySpecFromStruct(Y{})

	if bodySpec["original_value"].(reflect.Type).Kind() != reflect.Int {
		t.Errorf("Expected type of original_value was int, got %s", bodySpec["original_value"])
	}

	if bodySpec["value_plus_one"].(reflect.Type).Kind() != reflect.Int32 {
		t.Errorf("Expected type of value_plus_one was int32, got %s", bodySpec["value_plus_one"])
	}

	if bodySpec["str_arg"].(reflect.Type).Kind() != reflect.String {
		t.Errorf("Expected type of str_arg was string, got %s", bodySpec["str_arg"])
	}

	if bodySpec["bool_arg"].(reflect.Type).Kind() != reflect.Bool {
		t.Errorf("Expected type of bool_arg was bool, got %s", bodySpec["bool_arg"])
	}

	if bodySpec["struct_arg"].(BodySpecMap)["float_arg"].(reflect.Type).Kind() != reflect.Float32 {
		t.Errorf("Expected type of struct_arg/float_arg was float32, got %s", bodySpec["struct_arg"].(BodySpecMap)["float_arg"])
	}
}
