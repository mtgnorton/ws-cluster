package kit

import (
	"testing"
)

func TestMax(t *testing.T) {
	tests := []struct {
		name       string
		a, b, want int
	}{
		{"正整数", 5, 3, 5},
		{"负整数", -2, -5, -2},
		{"相等值", 7, 7, 7},
		{"零和正数", 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Max(tt.a, tt.b); got != tt.want {
				t.Errorf("Max(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name       string
		a, b, want int
	}{
		{"正整数", 5, 3, 3},
		{"负整数", -2, -5, -5},
		{"相等值", 7, 7, 7},
		{"零和负数", 0, -10, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Min(tt.a, tt.b); got != tt.want {
				t.Errorf("Min(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestRound(t *testing.T) {
	testCases := []struct {
		input    float64
		decimals int
		expected float64
	}{
		{3.14159, 2, 3.14},
		{3.14159, 3, 3.142},
		{3.14159, 4, 3.1416},
		{-3.14159, 2, -3.14},
		{0.0, 2, 0.0},
		{1.23456789, 5, 1.23457},
	}

	for _, tc := range testCases {
		result := RoundF64(tc.input, tc.decimals)
		if result != tc.expected {
			t.Errorf("Round(%f, %d) = %f; 期望 %f", tc.input, tc.decimals, result, tc.expected)
		}
	}
}
