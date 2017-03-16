package porthos

import "testing"

func TestUnmarshalJSONTo(t *testing.T) {
	type Test struct {
		A int `json:"a"`
	}

	res := Response{
		ContentType: "application/json",
		Content:     []byte("{\"a\": 1}"),
	}

	var test Test
	err := res.UnmarshalJSONTo(&test)

	if err != nil {
		t.Error("Error with json response:", err)
	}

	if test.A != 1 {
		t.Errorf("Expected json value was 1, got %d", test.A)
	}
}

func TestUnmarshalJSONToInvalidContentType(t *testing.T) {
	type Test struct {
		A int `json:"a"`
	}

	res := Response{
		ContentType: "text/plain",
		Content:     []byte("{\"a\": 1}"),
	}

	var test Test
	err := res.UnmarshalJSONTo(&test)

	if err == nil || err.Error() != "The content type of the response is not 'application/json'" {
		t.Error("Content type error was expected")
	}
}

func TestUnmarshalJSON(t *testing.T) {
	res := Response{
		ContentType: "application/json",
		Content:     []byte("{\"a\": 1.0}"),
	}

	json, err := res.UnmarshalJSON()

	if err != nil {
		t.Error("Error with json response:", err)
	}

	if json["a"] != 1.0 {
		t.Errorf("Expected json value was 1, got %d", json["a"])
	}
}

func TestUnmarshalJSONInvalidContentType(t *testing.T) {
	res := Response{
		ContentType: "text/plain",
		Content:     []byte("{\"a\": 1.0}"),
	}

	_, err := res.UnmarshalJSON()

	if err == nil || err.Error() != "The content type of the response is not 'application/json'" {
		t.Error("Content type error was expected")
	}
}
