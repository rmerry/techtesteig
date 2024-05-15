package btcore

import (
	"testing"
)

func TestChecksum(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint32
	}{
		{
			name:     "Empty payload",
			input:    []byte{},
			expected: emptyPayloadChecksum,
		},
		{
			name:     "Non-empty payload",
			input:    []byte("hello world"),
			expected: 0xb8d462bc,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checksum(tt.input)
			if result != tt.expected {
				t.Errorf("checksum(%q) = 0x%x, expected 0x%x", tt.input, result, tt.expected)
			}
		})
	}
}
