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
	flags.Int("bar", 3, "the `quantity` of bars you want")
	flags.Float64("blegga", 0.1, "on the spectrum of `fondue`, where do you fall?")
	flags.String("baz", "", "do you want baz with that?")
	flags.String("qux", "\t\n", "rhymes with `ducks`")
	return &flags
}

var subCmdApp = App{"foo", "maybe foo, maybe bar", "1.0", true}
var singleCmdApp = App{"bar", "maybe bar, maybe baz", "1.1", false}

func TestUsageGlobal(t *testing.T) {
	cmds := []*command.Cmd{
		&command.Cmd{Name: "foo", Summary: "foo the thing"},
		&command.Cmd{Name: "barzywarzyflarzy", Summary: "bar the thing"},
		&command.Cmd{Name: "baz", Summary: "baz the thing"},
	}

	buf := new(bytes.Buffer)
	usage := newUsage(subCmdApp, buf)
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
	--bar=quantity	(default: 3)		the quantity of bars you want
	--baz=string	(default: "")		do you want baz with that?
	--blegga=fondue	(default: 0.1)		on the spectrum of fondue, where do you fall?
	--foo		(default: false)	do the foo
	--qux=ducks	(default: "\t\n")	rhymes with ducks

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
	usage := newUsage(subCmdApp, buf)
	usage.Command(cmd)

	assert.Equal(t, buf.String(), `NAME:
	foo - foo the thing

USAGE:
	foo foo [FOO]

VERSION:
	1.0

DESCRIPTION:
	more deeply foo the thing

OPTIONS:
	--bar=quantity	(default: 3)		the quantity of bars you want
	--baz=string	(default: "")		do you want baz with that?
	--blegga=fondue	(default: 0.1)		on the spectrum of fondue, where do you fall?
	--foo		(default: false)	do the foo
	--qux=ducks	(default: "\t\n")	rhymes with ducks

For help on global options run "foo help"
`)

	buf = new(bytes.Buffer)
	usage = newUsage(singleCmdApp, buf)
	usage.Command(cmd)

	assert.Equal(t, buf.String(), `NAME:
	bar - foo the thing

USAGE:
	bar [FOO]

VERSION:
	1.1

DESCRIPTION:
	more deeply foo the thing

OPTIONS:
	--bar=quantity	(default: 3)		the quantity of bars you want
	--baz=string	(default: "")		do you want baz with that?
	--blegga=fondue	(default: 0.1)		on the spectrum of fondue, where do you fall?
	--foo		(default: false)	do the foo
	--qux=ducks	(default: "\t\n")	rhymes with ducks
`)
}
