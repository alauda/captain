package helmrequest

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestCRDParser(t *testing.T) {
	result, err := createCRDObject(helmRequestCRDYaml)
	assert.Equal(t, err, nil)
	t.Log(result.String())
}
