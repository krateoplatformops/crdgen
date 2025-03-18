package transpiler

import (
	"errors"
	"testing"
)

type stringer struct {
}

func (s stringer) String() string {
	return "stringer"
}

func TestStrval(t *testing.T) {

	tests := []struct {
		input    any
		expected string
	}{
		{"hello", "hello"},
		{[]byte("world"), "world"},
		{errors.New("error occurred"), "error occurred"},
		{stringer{}, "stringer"},
		{123, "123"},
	}

	for _, test := range tests {
		result := strval(test.input)
		if result != test.expected {
			t.Errorf("strval(%v) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestStrslice(t *testing.T) {
	tests := []struct {
		input    any
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]any{"a", 123, nil, errors.New("error")}, []string{"a", "123", "error"}},
		{[]int{1, 2, 3}, []string{"1", "2", "3"}},
		{nil, []string{}},
		{"single", []string{"single"}},
	}

	for _, test := range tests {
		result := strslice(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("strslice(%v) = %v; want %v", test.input, result, test.expected)
			continue
		}
		for i := range result {
			if result[i] != test.expected[i] {
				t.Errorf("strslice(%v) = %v; want %v", test.input, result, test.expected)
				break
			}
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		element  string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
	}

	for _, test := range tests {
		result := contains(test.slice, test.element)
		if result != test.expected {
			t.Errorf("contains(%v, %v) = %v; want %v", test.slice, test.element, result, test.expected)
		}
	}
}
