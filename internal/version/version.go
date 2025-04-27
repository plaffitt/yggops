package version

import (
	"fmt"
	"runtime"
)

var (
	Version    = ""
	CommitHash = ""
	BuildTime  = ""
)

func BuildVersion() string {
	if Version == "" {
		Version = "dev"
	}

	if CommitHash == "" || BuildTime == "" {
		return fmt.Sprintf("%s %s", Version, runtime.Version())
	}

	return fmt.Sprintf("%s#%s (%s) %s", Version, CommitHash, BuildTime, runtime.Version())
}
