package regex

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	// DefaultResourceNameRegex is resource name regex for most kubernetes resources
	DefaultResourceNameRegex = `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
)

//IsValidResourceName validate if a resource name is valid
func IsValidResourceName(name string) bool {
	matched, _ := regexp.MatchString(DefaultResourceNameRegex, name)
	return matched
}

// DefaultResourceNameRegexError create kubernetes like error message in kubectl about invalid HelmRequest
// resource
func DefaultResourceNameRegexError(kind, name, key, value string) error {
	template := `The %s "%s" is invalid: %s: Invalid value: "%s": ` +
		`a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', ` +
		`and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '%s')`
	return errors.New(fmt.Sprintf(template, kind, name, key, value, DefaultResourceNameRegex))
}
