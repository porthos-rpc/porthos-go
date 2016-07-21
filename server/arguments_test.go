package server

import "testing"

func TestStringArgument(t *testing.T) {
    value := (&Argument{"text"}).AsString()

    if value != "text" {
        t.Errorf("Argument cast failed, expected: 'text', got: '%s'", value)
    }
}

func TestIntArgument(t *testing.T) {
    value := (&Argument{12}).AsInt()

    if value != 12 {
        t.Errorf("Argument cast failed, expected: '12', got: '%d'", value)
    }
}
