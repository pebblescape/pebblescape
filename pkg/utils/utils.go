package utils

import (
	"regexp"
)

// AppNamePattern specifies accept format for app names.
var AppNamePattern = regexp.MustCompile(`^[a-z\d]+(-[a-z\d]+)*$`)
