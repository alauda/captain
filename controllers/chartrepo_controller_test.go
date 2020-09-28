package controllers

import (
	"github.com/alauda/captain/tests"
	"github.com/bmizerany/assert"
	"testing"
)

func TestCompareChart(t *testing.T) {
	old := tests.LoadChart("../tests/fixtures/chart-old.yaml")
	n := tests.LoadChart("../tests/fixtures/chart-new.yaml")

	result := compareChart(*old, n)
	assert.Equal(t, result, true)

	result = compareChart(*old, old)
	assert.Equal(t, result, false)

}
