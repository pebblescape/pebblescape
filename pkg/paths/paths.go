package paths

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/kardianos/osext"
)

// SelfPath returns the location of currently running executable.
func SelfPath() string {
	filename, _ := osext.Executable()

	return filename
}
