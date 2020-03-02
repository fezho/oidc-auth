// Package version contains version information for this app which is set by the build scripts.
package version

import (
	"fmt"
	"os"
	"runtime"
)

var (
	// Version shows the version of auth-service.
	Version = "Not provided."
	// GitSHA shows the git commit id of auth-service.
	GitSHA = "Not provided."
	// Built shows the built time of the binary.
	Built = "Not provided."
)

// PrintVersionAndExit prints versions from the array returned by Info() and exit
func PrintVersionAndExit() {
	for _, i := range Info() {
		fmt.Printf("%v\n", i)
	}
	os.Exit(0)
}

// Info returns an array of various service versions
func Info() []string {
	return []string{
		fmt.Sprintf("Version:    %s", Version),
		fmt.Sprintf("Git SHA:    %s", GitSHA),
		fmt.Sprintf("Built At:   %s", Built),
		fmt.Sprintf("Go Version: %s", runtime.Version()),
		fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
