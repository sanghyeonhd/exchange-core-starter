package decimal

import "testing"

func TestParseAndString(t *testing.T) {
	tests := map[string]int64{
		"50000.00": 5000000,
		"0.01":     1,
		"1":        100,
		"-1.23":    -123,
	}

	for input, want := range tests {
		got, err := Parse(input, 2)
		if err != nil {
			t.Fatalf("Parse(%q): %v", input, err)
		}
		if got.Value != want {
			t.Fatalf("Parse(%q) value = %d, want %d", input, got.Value, want)
		}
		if got.String() != input && input != "1" {
			t.Fatalf("String() = %q, want %q", got.String(), input)
		}
	}
}

func TestParseRejectsPrecisionLoss(t *testing.T) {
	if _, err := Parse("1.234", 2); err == nil {
		t.Fatal("expected precision loss rejection")
	}
}

func TestAddScaleMismatch(t *testing.T) {
	_, err := New(1, 2).Add(New(1, 6))
	if err != ErrScaleMismatch {
		t.Fatalf("err = %v, want %v", err, ErrScaleMismatch)
	}
}
