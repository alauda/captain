// +build go1.9

package assert

import "testing"

// Mark marks a function as a testing helper. Works only on Go 1.9 and above.
// See https://golang.org/pkg/testing/#T.Helper for more information.
var Mark = (*testing.T).Helper
