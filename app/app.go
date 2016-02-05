// The app package provides a simple App struct to describe a command-line
// application, and a Usage interface, which describers the global and
// command-specific usage of the App.
package app

import "os"

// A simple representation of a command-line application
type App struct {
	Name        string // the binary name of the application
	Description string // a short description of what the application does
	Version     string // the current version of the application
}

// Usage produces the default implementation of Usage for this App, which
// prints tab-formatted output to STDOUT.
func (a App) Usage() Usage {
	return newUsage(a, os.Stdout)
}
