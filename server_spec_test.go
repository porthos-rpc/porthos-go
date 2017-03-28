package porthos

import (
	"testing"
)

func TestBodySpecSimple(t *testing.T) {
	type X struct {
		Value int `json:"value"`
	}

	bodySpec := BodySpecFromStruct(X{})

	if bodySpec["value"].(string) != "int" {
		t.Errorf("Expected type of value was int, got %s", bodySpec["value"])
	}
}

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

	if bodySpec["original_value"].(string) != "int" {
		t.Errorf("Expected type of original_value was int, got %s", bodySpec["original_value"])
	}

	if bodySpec["value_plus_one"].(string) != "int32" {
		t.Errorf("Expected type of value_plus_one was int32, got %s", bodySpec["value_plus_one"])
	}

	if bodySpec["str_arg"].(string) != "string" {
		t.Errorf("Expected type of str_arg was string, got %s", bodySpec["str_arg"])
	}

	if bodySpec["bool_arg"].(string) != "bool" {
		t.Errorf("Expected type of bool_arg was bool, got %s", bodySpec["bool_arg"])
	}

	if bodySpec["struct_arg"].(BodySpecMap)["float_arg"].(string) != "float32" {
		t.Errorf("Expected type of struct_arg/float_arg was float32, got %s", bodySpec["struct_arg"].(BodySpecMap)["float_arg"])
	}
}
