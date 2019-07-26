// Package assert introduces a bunch of helpers to simplify the tests writing,
// assuming you are using the default testing package.
package assert

import (
	"testing"
)

var assertion Assertion

// Equal is the global version of Assertion.Equal.
func Equal(t *testing.T, expected, actual interface{}) Assertion {
	Mark(t)

	return assertion.Equal(t, expected, actual)
}

// NotEqual is the global version of Assertion.Equal.
func NotEqual(t *testing.T, expected, actual interface{}) Assertion {
	Mark(t)

	return assertion.NotEqual(t, expected, actual)
}

// True is the global version of Assertion.Equal.
func True(t *testing.T, cond bool) Assertion {
	Mark(t)

	return assertion.True(t, cond)
}

// False is the global version of Assertion.Equal.
func False(t *testing.T, cond bool) Assertion {
	Mark(t)

	return assertion.False(t, cond)
}

// Nil is the global version of Assertion.Nil.
func Nil(t *testing.T, v interface{}) Assertion {
	Mark(t)

	return assertion.Nil(t, v)
}

// NotNil is the global version of Assertion.NotNil.
func NotNil(t *testing.T, v interface{}) Assertion {
	Mark(t)

	return assertion.NotNil(t, v)
}

// Error is the global version of Assertion.Error.
func Error(t *testing.T, err error, message ...string) Assertion {
	Mark(t)

	return assertion.Error(t, err, message...)
}

// Len is the global version of Assertion.Len.
func Len(t *testing.T, length int, v interface{}) Assertion {
	Mark(t)

	return assertion.Len(t, length, v)
}

// Panic is the global version of Assertion.Panic.
func Panic(t *testing.T, fn func()) Assertion {
	Mark(t)

	return assertion.Panic(t, fn)
}
