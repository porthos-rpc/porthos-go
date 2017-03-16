package porthos

import (
	"testing"
)

func TestCallWithDefaultTimeout(t *testing.T) {
	c := newCall(&Client{defaultTTL: 10}, "doSomething")
	timeout := c.getTimeout()

	if timeout != 10 {
		t.Errorf("Got an unexpected timeout: %d", timeout)
	}
}

func TestCallWithSpecificTimeout(t *testing.T) {
	c := newCall(&Client{defaultTTL: 10}, "doSomething").WithTimeout(20)
	timeout := c.getTimeout()

	if timeout != 20 {
		t.Errorf("Got an unexpected timeout: %d", timeout)
	}
}

func TestCallWithBody(t *testing.T) {
	c := newCall(&Client{}, "doSomething").WithBody([]byte("test123"))

	if string(c.body) != "test123" {
		t.Errorf("Got an unexpected body: %s", string(c.body))
	}
}

func TestCallWithArgs(t *testing.T) {
	c := newCall(&Client{}, "doSomething").WithArgs(10, 20)

	if string(c.body) != "[10,20]" {
		t.Errorf("Got an unexpected body: %s", string(c.body))
	}
}

func TestCallWithMap(t *testing.T) {
	c := newCall(&Client{}, "doSomething").WithMap(map[string]interface{}{"fieldX": 5, "fieldY": 10})

	if string(c.body) != `{"fieldX":5,"fieldY":10}` {
		t.Errorf("Got an unexpected body: %s", string(c.body))
	}
}
