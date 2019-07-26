package assert

import "testing"

// EQ is an alias for Assertion.Equal.
func (a Assertion) EQ(t *testing.T, expected, actual interface{}) Assertion {
	Mark(t)

	return a.Equal(t, expected, actual)
}

// EQ is the global version of Assertion.EQ.
var EQ = Equal

// NEQ is an alias for Assertion.NotEqual.
func (a Assertion) NEQ(t *testing.T, expected, actual interface{}) Assertion {
	Mark(t)

	return a.NotEqual(t, expected, actual)
}

// NEQ is an alias for Assertion.NEQ.
var NEQ = NotEqual

// OK is an alias for Assertion.True.
func (a Assertion) OK(t *testing.T, assertion bool) Assertion {
	Mark(t)

	return a.True(t, assertion)
}

// OK is the global version of Assertion.OK.
var OK = True

// Present is an alias for Assertion.NotNil.
func (a Assertion) Present(t *testing.T, v interface{}) Assertion {
	Mark(t)

	return a.NotNil(t, v)
}

// Present is the global version of Assertion.Present.
var Present = NotNil

// Err is an alias for Assertion.Error.
func (a Assertion) Err(t *testing.T, err error, message ...string) Assertion {
	Mark(t)

	return a.Error(t, err, message...)
}

// Err is the global version of Assertion.Err.
var Err = Error
