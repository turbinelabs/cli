package app

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/turbinelabs/cli/command"
	tbnflag "github.com/turbinelabs/nonstdlib/flag"
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

func testFlagsFromEnv(scopes ...string) tbnflag.FromEnv {
	fe := tbnflag.NewFromEnv(testFlags(), scopes...)
	os.Setenv(fe.Prefix()+"FOO", "bar")
	os.Setenv(fe.Prefix()+"BAZ", "blegga")
	fe.Fill()
	fmt.Println(fe.Prefix())
	fmt.Println(fe.Filled())
	return fe
}

var subCmdApp = App{"foo", "maybe foo, maybe bar", "1.0", true}
var singleCmdApp = App{"bar", "maybe bar, maybe baz", "1.1", false}

func TestUsageGlobal(t *testing.T) {
	cmds := []*command.Cmd{
		{Name: "foo", Summary: "foo the thing"},
		{Name: "barzywarzyflarzy", Summary: "bar the thing"},
		{Name: "baz", Summary: "baz the thing"},
	}

	buf := new(bytes.Buffer)
	usage := newUsage(subCmdApp, buf, 84, true)
	usage.Global(cmds, testFlagsFromEnv(subCmdApp.Name))

	assert.Equal(t, buf.String(), bold("NAME")+`
    foo - maybe foo, maybe bar

`+bold("USAGE")+`
    foo [GLOBAL OPTIONS] <command> [COMMAND OPTIONS] [arguments...]

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

    Global options can also be configured via upper-case, underscore-delimeted
    environment variables prefixed with "FOO_". For example, "--some-flag"
    becomes "FOO_SOME_FLAG". Command-line flags take precedence over environment
    variables.

    Options currently configured from the Environment:

        FOO_BAZ=blegga
        FOO_FOO=bar

Run "foo help <command>" for more details on a specific command.

`)

	buf = new(bytes.Buffer)
	usage = newUsage(subCmdApp, buf, 24, true)
	usage.Global(cmds, testFlagsFromEnv(subCmdApp.Name))

	assert.Equal(t, buf.String(), bold("NAME")+`
    foo - maybe foo,
    maybe bar

`+bold("USAGE")+`
    foo [GLOBAL
    OPTIONS]
    <command>
    [COMMAND
    OPTIONS]
    [arguments...]

`+bold("VERSION")+`
    1.0

`+bold("COMMANDS")+`
    `+ul("foo")+`
            foo the
            thing

    `+ul("barzywarzyflarzy")+`
            bar the
            thing

    `+ul("baz")+`
            baz the
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

    Global options
    can also be
    configured via
    upper-case,
    underscore-delimeted
    environment
    variables
    prefixed with
    "FOO_". For
    example,
    "--some-flag"
    becomes
    "FOO_SOME_FLAG".
    Command-line
    flags take
    precedence over
    environment
    variables.

    Options
    currently
    configured from
    the Environment:

        FOO_BAZ=blegga
        FOO_FOO=bar

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
	usage := newUsage(subCmdApp, buf, 84, true)
	usage.Command(
		cmd,
		testFlagsFromEnv(subCmdApp.Name),
		testFlagsFromEnv(subCmdApp.Name, cmd.Name),
	)

	assert.Equal(t, buf.String(), bold("NAME")+`
    foo - foo the thing

`+bold("USAGE")+`
    foo [GLOBAL OPTIONS] foo [FOO]

`+bold("VERSION")+`
    1.0

`+bold("DESCRIPTION")+`
    more `+bold("deeply")+` foo the `+ul("thing")+`!

    pay attention to explicit newlines, and pay special attention to this
    pseudo-code:

        if (happy) {
            dance()
        }

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

    Global options can also be configured via upper-case, underscore-delimeted
    environment variables prefixed with "FOO_". For example, "--some-flag"
    becomes "FOO_SOME_FLAG". Command-line flags take precedence over environment
    variables.

    Options currently configured from the Environment:

        FOO_BAZ=blegga
        FOO_FOO=bar

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

    Options can also be configured via upper-case, underscore-delimeted
    environment variables prefixed with "FOO_FOO_". For example, "--some-flag"
    becomes "FOO_FOO_SOME_FLAG". Command-line flags take precedence over
    environment variables.

    Options currently configured from the Environment:

        FOO_FOO_BAZ=blegga
        FOO_FOO_FOO=bar

`)

	buf = new(bytes.Buffer)
	usage = newUsage(singleCmdApp, buf, 84, true)
	usage.Command(
		cmd,
		testFlagsFromEnv(singleCmdApp.Name),
		testFlagsFromEnv(singleCmdApp.Name),
	)

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

    Options can also be configured via upper-case, underscore-delimeted
    environment variables prefixed with "BAR_". For example, "--some-flag"
    becomes "BAR_SOME_FLAG". Command-line flags take precedence over environment
    variables.

    Options currently configured from the Environment:

        BAR_BAZ=blegga
        BAR_FOO=bar

`)

	buf = new(bytes.Buffer)
	usage = newUsage(singleCmdApp, buf, 24, true)
	usage.Command(
		cmd,
		testFlagsFromEnv(singleCmdApp.Name),
		testFlagsFromEnv(singleCmdApp.Name),
	)

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

    Options can also
    be configured
    via upper-case,
    underscore-delimeted
    environment
    variables
    prefixed with
    "BAR_". For
    example,
    "--some-flag"
    becomes
    "BAR_SOME_FLAG".
    Command-line
    flags take
    precedence over
    environment
    variables.

    Options
    currently
    configured from
    the Environment:

        BAR_BAZ=blegga
        BAR_FOO=bar

`)
}
