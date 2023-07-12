package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArrayContainsStr(t *testing.T) {
	// arrange
	in := []string{"a", "b", "c"}
	// act
	has := ArrayContainsStr(in, "b")
	// assert
	require.Equal(t, true, has)

	has = ArrayContainsStr(in, "d")
	require.Equal(t, false, has)
}
