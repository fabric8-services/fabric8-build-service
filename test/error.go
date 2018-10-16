package test

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ErrorCheck type
type ErrorCheck func(t require.TestingT, actualError error)

// AssertError Assert multiple errors if there is an error
func AssertError(t require.TestingT, actualError error, checks ...ErrorCheck) {
	require.Error(t, actualError)
	for _, check := range checks {
		check(t, actualError)
	}
}

// IsOfType helper for assert.IsOfType
func IsOfType(expectedType interface{}) ErrorCheck {
	return func(t require.TestingT, actualError error) {
		assert.IsType(t, expectedType, errors.Cause(actualError))
	}
}

// HasMessage check if we have message
func HasMessage(expectedMsg string, args ...interface{}) ErrorCheck {
	return func(t require.TestingT, actualError error) {
		assert.Equal(t, fmt.Sprintf(expectedMsg, args...), actualError.Error())
	}
}

// HasMessageContaining check if message contains
func HasMessageContaining(expectedMsg string, args ...interface{}) ErrorCheck {
	return func(t require.TestingT, actualError error) {
		assert.Contains(t, actualError.Error(), fmt.Sprintf(expectedMsg, args...))
	}
}
