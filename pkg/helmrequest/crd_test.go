package helmrequest

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestCRDParser(t *testing.T) {
	result, err := CreateCRDObject(helmRequestCRDYaml)
	assert.Equal(t, err, nil)
	t.Log(result.String())
}
