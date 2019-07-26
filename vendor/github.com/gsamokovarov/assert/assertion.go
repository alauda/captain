package assert

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

// Assertion contains all the assertions functions with chaining support.
type Assertion struct{}

// Equal tests two objects for equality.
func (a Assertion) Equal(t *testing.T, expected, actual interface{}) Assertion {
	Mark(t)

	if isNil(expected) || isNil(actual) {
		if isNil(expected) && isNil(actual) {
			return a
		}

		Diff(t, true, expected, actual)
	}

	val := reflect.ValueOf(expected)
	typ := reflect.TypeOf(actual)

	if !val.Type().ConvertibleTo(typ) {
		t.Fatalf("Cannot compare %v with %v", val.Type(), typ)
	}

	eval := val.Convert(typ).Interface()

	// Check for NaN. NaN is the only value that is not equal to itself.
	// That's why all the drama.
	if eval, ok := eval.(float64); ok {
		if actual := actual.(float64); ok {
			if math.IsNaN(eval) && math.IsNaN(actual) {
				return a
			}
		}
	}

	if !reflect.DeepEqual(eval, actual) {
		Diff(t, true, eval, actual)
	}

	return a
}

// NotEqual tests two objects for inequality.
func (a Assertion) NotEqual(t *testing.T, expected, actual interface{}) Assertion {
	Mark(t)

	// Shortcut the nil check by abusing Go's == nil. This will catch early any
	// nil assertion early. Be it the literal nil value or the zero value of a
	// referential type.
	if isNil(expected) || isNil(actual) {
		if isNil(expected) && isNil(actual) {
			Diff(t, false, expected, actual)
		}

		return a
	}

	typ := reflect.TypeOf(actual)
	val := reflect.ValueOf(expected)

	if !val.Type().ConvertibleTo(typ) {
		t.Fatalf("Cannot compare %v with %v", val.Type(), typ)
	}

	eval := val.Convert(typ).Interface()

	// Check for NaN. NaN is the only value that is not equal to itself.
	// That's why all the drama.
	if eval, ok := eval.(float64); ok {
		if actual := actual.(float64); ok {
			if math.IsNaN(eval) && math.IsNaN(actual) {
				Diff(t, false, eval, actual)
			}

			return a
		}
	}

	if reflect.DeepEqual(eval, actual) {
		Diff(t, false, eval, actual)
	}

	return a
}

// True fails the current test if the assertion is false.
func (a Assertion) True(t *testing.T, cond bool) Assertion {
	Mark(t)

	return Equal(t, true, cond)
}

// False fails the current test if the assertion is true.
func (a Assertion) False(t *testing.T, cond bool) Assertion {
	Mark(t)

	return a.Equal(t, false, cond)
}

// Nil fails the current test if the values is not nil.
func (a Assertion) Nil(t *testing.T, v interface{}) Assertion {
	Mark(t)

	return a.Equal(t, nil, v)
}

// NotNil fails the current test if the values is nil.
func (a Assertion) NotNil(t *testing.T, v interface{}) Assertion {
	Mark(t)

	return a.NotEqual(t, nil, v)
}

// Error fails the current test if the values is nil error, or it's Error()
// string does not match the optional message. The message can be given in
// parts that would be joined by the empty string.
func (a Assertion) Error(t *testing.T, err error, message ...string) Assertion {
	Mark(t)

	a.NotNil(t, err)
	if len(message) != 0 {
		a.Equal(t, strings.Join(message, ""), err.Error())
	}

	return a
}

// Len fails the current test if the value doesn't have the expected length.
// Only arrays, chans, maps, slices and strings can have length calculated.
func (a Assertion) Len(t *testing.T, length int, v interface{}) Assertion {
	Mark(t)

	val := reflect.Indirect(reflect.ValueOf(v))

	switch val.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		a.Equal(t, length, val.Len())
	default:
		t.Fatalf("Cannot get the length of %v", val)
	}

	return a
}

// Panic fails the current test if the given function does not panic.
func (a Assertion) Panic(t *testing.T, fn func()) Assertion {
	Mark(t)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Expected a panic")
		}
	}()

	fn()

	return a
}
