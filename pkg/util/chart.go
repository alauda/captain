package util

import "strings"

// ParseChartName is a simple function that parse chart name
func ParseChartName(name string) (repo, chart string) {
	data := strings.Split(name, "/")
	if len(data) == 1 {
		return "", name
	}
	return data[0], data[1]
}
