package util

import (
	"testing"
)

func TestIntToHex(t *testing.T) {
	tests := []struct {
		input    int64
		expected []byte
	}{
		{0, []byte("0")},
		{10, []byte("a")},
		{255, []byte("ff")},
		{1024, []byte("400")},
		{-1, []byte("-1")},
	}

	for _, test := range tests {
		result := IntToHex(test.input)
		if string(result) != string(test.expected) {
			t.Errorf("IntToHex(%d) = %s; expected %s", test.input, result, test.expected)
		}
	}
}
