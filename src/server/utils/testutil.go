package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// CheckError calls f and asserts it did not return an error
func CheckError(t *testing.T, f func() error) {
	t.Helper()
	err := f()
	require.NoError(t, err)
}
