package util

import "strings"

func ParseChartName(name string) (repo, chart string) {
	data := strings.Split(name, "/")
	if len(data) == 1 {
		return "", name
	}
	return data[0], data[1]
}
