package utils

import (
	"regexp"
)

var AppNamePattern = regexp.MustCompile(`^[a-z\d]+(-[a-z\d]+)*$`)
