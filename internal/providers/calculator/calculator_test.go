package calculator

import (
	"math"
	"testing"
)

func TestEvaluate(t *testing.T) {
	cases := []struct {
		expr     string
		expected float64
	}{
		{"2 + 2", 4},
		{"10 - 3 * 2", 4},
		{"(10 - 3) * 2", 14},
		{"2 ^ 3", 8},
		{"sqrt(16)", 4},
		{"abs(-42)", 42},
		{"100 / 4 + 5 * 2", 35},
	}

	for _, c := range cases {
		val, err := Evaluate(c.expr)
		if err != nil {
			t.Fatalf("Evaluate(%q) returned error: %v", c.expr, err)
		}
		if math.Abs(val-c.expected) > 1e-6 {
			t.Errorf("Evaluate(%q) = %v, expected %v", c.expr, val, c.expected)
		}
	}
}
