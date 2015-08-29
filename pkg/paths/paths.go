package paths

import (
	"github.com/pebblescape/pebblescape/Godeps/_workspace/src/github.com/kardianos/osext"
)

func SelfPath() string {
	filename, _ := osext.Executable()

	return filename
}
