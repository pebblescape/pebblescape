package utils

import (
	"errors"
	"regexp"
)

var AppNamePattern = regexp.MustCompile(`^[a-z\d]+(-[a-z\d]+)*$`)
var AppNameError = errors.New("App name contains invalid characters. Alphanumeric only, no whitespace.")
