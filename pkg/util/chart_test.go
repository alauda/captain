package util

import (
	"testing"

	"github.com/gsamokovarov/assert"
)

func TestParseChart(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		repo, name := ParseChartName("stable/gitlab")
		assert.Equal(t, repo, "stable")
		assert.Equal(t, name, "gitlab")
	})

}
