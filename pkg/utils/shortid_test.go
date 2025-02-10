package utils_test

import (
	"testing"

	"github.com/ivaeg3/url-shortener/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		id       uint64
		expected int
	}{
		{0, 10},
		{1, 10},
		{100, 10},
		{1000000, 10},
		{18446744073709551615, 10}, // Max uint64
	}

	for _, tt := range tests {
		t.Run("Encode test", func(t *testing.T) {
			result := utils.Encode(tt.id)
			assert.Len(t, result, tt.expected, "Encode(%d) = %s, expected length %d", tt.id, result, tt.expected)
		})
	}
}
