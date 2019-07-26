// +build !go1.9

package assert

import "testing"

// Mark marks a function as a testing helper. Works only on Go 1.9 and above.
// See https://golang.org/pkg/testing/#T.Helper for more information. Still,
// it's recommended to call it, even if it's a noop for Go 1.8 and below.
func Mark(t *testing.T) {}
