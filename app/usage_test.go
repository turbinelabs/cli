package app

import (
	"bytes"
	"flag"
	"testing"

	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/test/assert"
)

func testFlags() *flag.FlagSet {
	var flags flag.FlagSet
	flags.Bool("foo", false, "do the foo")
	flags.Int("bar", 3, "the `quantity` of bars you want")
	flags.Float64("blegga", 0.1, "on the spectrum of `fondue`, where do you fall?")
	flags.String("baz", "", "do you want baz with that?")
	flags.String("qux", "\t\n", "rhymes with `ducks`")
	flags.String("fnord", "", "rhymes with `fjord`")
	return &flags
}

var subCmdApp = App{"foo", "maybe foo, maybe bar", "1.0", true}
var singleCmdApp = App{"bar", "maybe bar, maybe baz", "1.1", false}
var flagsFromEnv = map[string]string{"FOO": "bar", "BAZ": "blegga"}

func TestUsageGlobal(t *testing.T) {
	cmds := []*command.Cmd{
		{Name: "foo", Summary: "foo the thing"},
		{Name: "barzywarzyflarzy", Summary: "bar the thing"},
		{Name: "baz", Summary: "baz the thing"},
	}

	buf := new(bytes.Buffer)
	usage := newUsage(subCmdApp, buf, 84)
	usage.Global(cmds, testFlags(), flagsFromEnv)

	assert.Equal(t, buf.String(), bold("NAME")+`
    foo - maybe foo, maybe bar

`+bold("USAGE")+`
    foo [global options] <command> [command options] [arguments...]

`+bold("VERSION")+`
    1.0

`+bold("COMMANDS")+`
    `+ul("foo")+`     foo the thing

    `+ul("barzywarzyflarzy")+`
            bar the thing

    `+ul("baz")+`     baz the thing

`+bold("GLOBAL OPTIONS")+`
    --`+ul("bar")+`=quantity
            (default: 3)
            the quantity of bars you want

    --`+ul("baz")+`=string
            do you want baz with that?

    --`+ul("blegga")+`=fondue
            (default: 0.1)
            on the spectrum of fondue, where do you fall?

    --`+ul("fnord")+`=fjord
            rhymes with fjord

    --`+ul("foo")+`   (default: false)
            do the foo

    --`+ul("qux")+`=ducks
            (default: "\t\n")
            rhymes with ducks

Global options can also be configured via upper-case environment variables
prefixed with "FOO_" For example, "--some-flag" --> "FOO_SOME_FLAG".
Command-line flags take precidence over environment variables. Options currently
configured from the Environment:

    BAZ=blegga
    FOO=bar

Run "foo help <command>" for more details on a specific command.
`)

	buf = new(bytes.Buffer)
	usage = newUsage(subCmdApp, buf, 24)
	usage.Global(cmds, testFlags(), flagsFromEnv)

	assert.Equal(t, buf.String(), bold("NAME")+`
    foo - maybe foo,
    maybe bar

`+bold("USAGE")+`
    foo [global
    options]
    <command>
    [command
    options]
    [arguments...]

`+bold("VERSION")+`
    1.0

`+bold("COMMANDS")+`
    `+ul("foo")+`     foo the
            thing

    `+ul("barzywarzyflarzy")+`
            bar the
            thing

    `+ul("baz")+`     baz the
            thing

`+bold("GLOBAL OPTIONS")+`
    --`+ul("bar")+`=quantity
            (default:
            3)
            the
            quantity
            of bars
            you want

    --`+ul("baz")+`=string
            do you
            want baz
            with
            that?

    --`+ul("blegga")+`=fondue
            (default:
            0.1)
            on the
            spectrum
            of
            fondue,
            where do
            you
            fall?

    --`+ul("fnord")+`=fjord
            rhymes
            with
            fjord

    --`+ul("foo")+`   (default:
            false)
            do the
            foo

    --`+ul("qux")+`=ducks
            (default:
            "\t\n")
            rhymes
            with
            ducks

Global options can
also be configured
via upper-case
environment
variables prefixed
with "FOO_" For
example,
"--some-flag" -->
"FOO_SOME_FLAG".
Command-line flags
take precidence over
environment
variables. Options
currently configured
from the
Environment:

    BAZ=blegga
    FOO=bar

Run "foo help
<command>" for more
details on a
specific command.
`)
}

func TestUsageCommand(t *testing.T) {
	desc := `more {{bold "deeply"}} foo the {{ul "thing"}}!



pay attention to explicit newlines,
and pay special attention to this pseudo-code:

    if (happy) {
        dance()
    }
`

	cmd := &command.Cmd{
		Name:        "foo",
		Summary:     "foo the thing",
		Description: desc,
		Usage:       "[FOO]",
		Flags:       *testFlags(),
	}

	buf := new(bytes.Buffer)
	usage := newUsage(subCmdApp, buf, 84)
	usage.Command(cmd, flagsFromEnv)

	assert.Equal(t, buf.String(), bold("NAME")+`
    foo - foo the thing

`+bold("USAGE")+`
    foo foo [FOO]

`+bold("VERSION")+`
    1.0

`+bold("DESCRIPTION")+`
    more `+bold("deeply")+` foo the `+ul("thing")+`!

    pay attention to explicit newlines, and pay special attention to this
    pseudo-code:

        if (happy) {
            dance()
        }

`+bold("OPTIONS")+`
    --`+ul("bar")+`=quantity
            (default: 3)
            the quantity of bars you want

    --`+ul("baz")+`=string
            do you want baz with that?

    --`+ul("blegga")+`=fondue
            (default: 0.1)
            on the spectrum of fondue, where do you fall?

    --`+ul("fnord")+`=fjord
            rhymes with fjord

    --`+ul("foo")+`   (default: false)
            do the foo

    --`+ul("qux")+`=ducks
            (default: "\t\n")
            rhymes with ducks

Options can also be configured via upper-case environment variables prefixed
with "FOO_" For example, "--some-flag" --> "FOO_SOME_FLAG". Command-line flags
take precidence over environment variables. Options currently configured from
the Environment:

    BAZ=blegga
    FOO=bar

For global options run "foo help".
`)

	buf = new(bytes.Buffer)
	usage = newUsage(singleCmdApp, buf, 84)
	usage.Command(cmd, flagsFromEnv)

	assert.Equal(t, buf.String(), bold("NAME")+`
    bar - foo the thing

`+bold("USAGE")+`
    bar [FOO]

`+bold("VERSION")+`
    1.1

`+bold("DESCRIPTION")+`
    more `+bold("deeply")+` foo the `+ul("thing")+`!

    pay attention to explicit newlines, and pay special attention to this
    pseudo-code:

        if (happy) {
            dance()
        }

`+bold("OPTIONS")+`
    --`+ul("bar")+`=quantity
            (default: 3)
            the quantity of bars you want

    --`+ul("baz")+`=string
            do you want baz with that?

    --`+ul("blegga")+`=fondue
            (default: 0.1)
            on the spectrum of fondue, where do you fall?

    --`+ul("fnord")+`=fjord
            rhymes with fjord

    --`+ul("foo")+`   (default: false)
            do the foo

    --`+ul("qux")+`=ducks
            (default: "\t\n")
            rhymes with ducks

Options can also be configured via upper-case environment variables prefixed
with "BAR_" For example, "--some-flag" --> "BAR_SOME_FLAG". Command-line flags
take precidence over environment variables. Options currently configured from
the Environment:

    BAZ=blegga
    FOO=bar
`)

	buf = new(bytes.Buffer)
	usage = newUsage(singleCmdApp, buf, 24)
	usage.Command(cmd, flagsFromEnv)

	assert.Equal(t, buf.String(), bold("NAME")+`
    bar - foo the
    thing

`+bold("USAGE")+`
    bar [FOO]

`+bold("VERSION")+`
    1.1

`+bold("DESCRIPTION")+`
    more
    `+bold("deeply")+`
    foo the
    `+ul("thing")+`!

    pay attention to
    explicit
    newlines, and
    pay special
    attention to
    this
    pseudo-code:

        if (happy) {
            dance()
        }

`+bold("OPTIONS")+`
    --`+ul("bar")+`=quantity
            (default:
            3)
            the
            quantity
            of bars
            you want

    --`+ul("baz")+`=string
            do you
            want baz
            with
            that?

    --`+ul("blegga")+`=fondue
            (default:
            0.1)
            on the
            spectrum
            of
            fondue,
            where do
            you
            fall?

    --`+ul("fnord")+`=fjord
            rhymes
            with
            fjord

    --`+ul("foo")+`   (default:
            false)
            do the
            foo

    --`+ul("qux")+`=ducks
            (default:
            "\t\n")
            rhymes
            with
            ducks

Options can also be
configured via
upper-case
environment
variables prefixed
with "BAR_" For
example,
"--some-flag" -->
"BAR_SOME_FLAG".
Command-line flags
take precidence over
environment
variables. Options
currently configured
from the
Environment:

    BAZ=blegga
    FOO=bar
`)
}
