// This package contains a trivial example use of the cli package
package cli_test

import (
	"fmt"

	"github.com/turbinelabs/tbn/cli"
	"github.com/turbinelabs/tbn/cli/app"
	"github.com/turbinelabs/tbn/cli/command"
)

// The typical pattern is to provide a public CmdXYZ() func for each
// sub-command you wish to provide. This function should initialize the
// command.Cmd, the command.Runner, and flags.
func CmdFoo() *command.Cmd {
	// typically the command.Runner is initialized as empty; all necessary
	// state should be provided via flags. One can inline the initializaton
	// of the command.Runner in the command.Cmd initialization if no flags
	// are necessary, but it's generally convenient to have a typed reference
	runner := new(fooRunner)

	cmd := &command.Cmd{
		Name:        "foo",
		Summary:     "Foo the bar",
		Usage:       "[OPTIONS] <bar>",
		Description: "Foo the bar until it's bazzy",
		Runner:      runner,
	}

	// The flag.FlagSet is a member of the command.Cmd, and the flag
	// value is a member of the command.Runner.
	cmd.Flags.StringVar(&runner.Baz, "baz", "baz", "The baz value.")

	return cmd
}

// The private command.Runner implementation should contain any state needed
// to execute the command. The values should be initialized via flags declared
// in the CmdXYZ() function.
type fooRunner struct {
	Baz string
}

// Run does the actual work, based on state provided by flags, and the
// args remaining after the flags have been parsed.
func (f *fooRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	// argument validation should occur at the top of the function, and
	// errors should be reported via the cmd.BadInput or cmd.BadInputf methods
	if len(args) < 1 {
		return cmd.BadInput("missing \"bar\" argument.")
	}
	bar := args[0]
	fmt.Printf("\n%s -> Foo-%s-%s\n\n", bar, bar, f.Baz)

	// In this case, there was no error. Errors should be returned via the
	// cmd.Error or cmd.Errorf methods.
	return command.NoError()
}

var blegga bool

func Example_main() {
	// declare your app.App
	app := app.App{
		Name:        "cli-example",
		Description: "an example CLI",
		Version:     "1.0.0",
	}

	// make a new CLI passing the app context and one or more sub commands
	// (in this case we have only CmdFoo)
	c := cli.NewCLI(
		app,
		CmdFoo(),
	)

	// Global flags can be used to modify global state
	c.Flags.BoolVar(&blegga, "blegga", true, "Should we blegga?")

	// run the Main function, which calls os.Exit with the appropriate exit status
	c.Main()
}
