package metadata

import (
	"fmt"
)

// Version represents the metadata identifying the running version of the CLI.
type Version struct {
	Commit    string
	Number    string
	GoOS      string
	GoVersion string
}

// String returns a formatted description of the CLI version, suitable for use
// in response to the `--version` flag.
func (v *Version) String() string {
	return fmt.Sprintf("BuildPulse Test Reporter %s (%s %s %s)\n", v.Number, v.GoOS, v.Commit, v.GoVersion)
}
