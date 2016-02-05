package app

import (
	"bytes"
	"flag"
	"testing"

	"github.com/turbinelabs/tbn/cli/command"
	"github.com/turbinelabs/tbn/testhelpers/assert"
)

func testFlags() *flag.FlagSet {
	var flags flag.FlagSet
	flags.Bool("foo", false, "do the foo")
	flags.Bool("bar", false, "harty har to the bar")
	return &flags
}

var testApp = App{"foo", "maybe foo, maybe bar", "1.0"}

func TestUsageGlobal(t *testing.T) {
	cmds := []*command.Cmd{
		&command.Cmd{Name: "foo", Summary: "foo the thing"},
		&command.Cmd{Name: "barzywarzyflarzy", Summary: "bar the thing"},
		&command.Cmd{Name: "baz", Summary: "baz the thing"},
	}

	buf := new(bytes.Buffer)
	usage := newUsage(testApp, buf)
	usage.Global(cmds, testFlags())

	assert.Equal(t, buf.String(), `NAME:
	foo - maybe foo, maybe bar

USAGE:
	foo [global options] <command> [command options] [arguments...]

VERSION:
	1.0

COMMANDS:
	foo			foo the thing
	barzywarzyflarzy	bar the thing
	baz			baz the thing

GLOBAL OPTIONS:
	--bar=false	harty har to the bar
	--foo=false	do the foo

Global options can also be configured via upper-case environment variables prefixed with "FOO"
For example, "--some_flag" --> "FOO_SOME_FLAG"

Run "foo help <command>" for more details on a specific command.
`)
}

func TestUsageCommand(t *testing.T) {
	cmd := &command.Cmd{
		Name:        "foo",
		Summary:     "foo the thing",
		Description: "more deeply foo the thing",
		Usage:       "[FOO]",
		Flags:       *testFlags(),
	}

	buf := new(bytes.Buffer)
	usage := newUsage(testApp, buf)
	usage.Command(cmd)

	assert.Equal(t, buf.String(), `NAME:
	foo - foo the thing

USAGE:
	foo foo [FOO]

DESCRIPTION:
	more deeply foo the thing

OPTIONS:
	--bar=false	harty har to the bar
	--foo=false	do the foo

For help on global options run "foo help"
`)
}
