package server

import (
	"encoding/json"
	"testing"
)

func TestStringArgument(t *testing.T) {
	value, _ := (&argument{"text"}).AsString()

	if value != "text" {
		t.Errorf("Argument cast failed, expected: 'text', got: '%s'", value)
	}
}

func TestIntArgument(t *testing.T) {
	value, _ := (&argument{json.Number("10")}).AsInt()

	if value != 10 {
		t.Errorf("Argument cast failed, expected: '10', got: '%d'", value)
	}
}

func TestBoolArgument(t *testing.T) {
	value, _ := (&argument{"true"}).AsBool()

	if value {
		t.Errorf("Argument cast failed, expected: 'true', got: 'false'")
	}
}

func TestInvalidIntArgument(t *testing.T) {
	value, err := (&argument{json.Number("xxx")}).AsInt()

	if err == nil {
		t.Errorf("Error as expected")
	}

	if value != 0 {
		t.Errorf("Argument cast returned non zero result, got: '%d'", value)
	}
}

func TestInvalidInt64Argument(t *testing.T) {
	value, err := (&argument{json.Number("64000000")}).AsInt()

	if err != nil {
		t.Errorf("Got an error", err)
	}

	if value != 64000000 {
		t.Errorf("Argument cast failed, expected: '64000000', got: '%d'", value)
	}
}

func TestInvalidBoolArgument(t *testing.T) {
	_, err := (&argument{"xxx"}).AsBool()

	if err != ErrTypeCast {
		t.Errorf("Expected ErrTypeCast, got %s", err)
	}
}
