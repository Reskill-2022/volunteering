package errors

import "testing"

func TestErrors(t *testing.T) {
	e1 := New("Base", 500)
	e2 := From(e1, "Mid", CodeFrom(e1))
	e3 := From(e2, "Top", CodeFrom(e2))

	if e3.Code != 500 {
		t.Errorf("expected code 500, got %d", e3.Code)
	}

	if got := e3.Error(); got != "Top: Mid: Base: " {
		t.Errorf("expected errors to unpack to 'Top: Mid: Base: ', got '%s'", got)
	}
}
