// The app package provides a simple App struct to describe a command-line
// application, and a Usage interface, which describers the global and
// command-specific usage of the App.
package app

import (
	"io"
	"os"
)

// A simple representation of a command-line application
type App struct {
	Name          string // the binary name of the application
	Description   string // a short description of what the application does
	VersionString string // the current version of the application
	HasSubCmds    bool   // whether or not the app has sub commands
}

// Usage produces the default implementation of Usage for this App, which
// prints tab-formatted output to STDOUT.
func (a App) Usage() Usage {
	return newUsage(a, os.Stdout, widthFromTerm, false)
}

// RedirectedUsage produces a Usage for this App, which prints
// tab-formatted output to the given Writer at a width of 80 columns.
func (a App) RedirectedUsage(writer io.Writer) Usage {
	return newUsage(a, writer, 80, false)
}

func (a App) Version() Version {
	return versionT{name: a.Name, version: a.VersionString, metadata: versionMetadata}
}
