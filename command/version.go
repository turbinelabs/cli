package command

import "fmt"

const VersionSummary = "Print the version and exit"

// Version produces a generic version Cmd which outputs the name
// and version of a command-line application to STDOUT.
func Version(name string, version string) *Cmd {
	return &Cmd{
		Name:        "version",
		Description: VersionSummary,
		Summary:     VersionSummary,
		Runner:      versionRunner{name, version},
	}
}

type versionRunner struct {
	name    string
	version string
}

func (v versionRunner) Run(cmd *Cmd, args []string) CmdErr {
	fmt.Println(v.name, "version", v.version)
	return NoError()
}
