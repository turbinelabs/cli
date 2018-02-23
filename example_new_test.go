/*
Copyright 2018 Turbine Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This package contains a trivial example use of the cli package
package cli_test

import (
	"fmt"
	"strconv"

	"github.com/turbinelabs/cli"
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/nonstdlib/flag/usage"
)

// The typical pattern is to provide a public Cmd() func. This function should
// initialize the command.Cmd, the command.Runner, and flags.
func Cmd() *command.Cmd {
	// typically the command.Runner is initialized only with internally-defined
	// state; all necessary external state should be provided via flags. One can
	// inline the initializaton of the command.Runner in the command.Cmd
	// initialization if no flags are necessary, but it's often convenient to
	// have a typed reference
	runner := &runner{}

	cmd := &command.Cmd{
		Name:        "adder",
		Summary:     "add a delimited string of integers together",
		Usage:       "[OPTIONS] <int>...",
		Description: "add a delimited string of integers together",
		Runner:      runner,
	}

	// The flag.FlagSet is a member of the command.Cmd, and the flag
	// value is a member of the command.Runner.
	cmd.Flags.BoolVar(&runner.verbose, "verbose", false, "Produce verbose output")

	// If we wrap flag.Required(...) around the usage string, Cmd.Run(...)
	// will fail if it is unspecified
	cmd.Flags.StringVar(&runner.thing, "thing", "", usage.Required("The thing"))

	return cmd
}

// The private command.Runner implementation should contain any state needed
// to execute the command. The values should be initialized via flags declared
// in the Cmd() function.
type runner struct {
	verbose bool
	thing   string
}

// Run does the actual work, based on state provided by flags, and the
// args remaining after the flags have been parsed.
func (f *runner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	ints := []int{}
	sum := 0
	// argument validation should occur at the top of the function, and
	// errors should be reported via the cmd.BadInput or cmd.BadInputf methods.
	// In this case, the main work of the function is done at the same time.
	for _, arg := range args {
		i, err := strconv.Atoi(arg)
		if err != nil {
			return cmd.BadInputf("Bad integer: \"%s\": %s", arg, err)
		}
		ints = append(ints, i)
		sum += i
	}

	if f.verbose && len(ints) > 0 {
		fmt.Print(ints[0])
		for _, i := range ints[1:] {
			fmt.Print(" + ", i)
		}
		fmt.Print(" = ")
	}

	fmt.Printf(`The thing: %s, the sum: %d`, f.thing, sum)

	// In this case, there was no error. Errors should be returned via the
	// cmd.Error or cmd.Errorf methods.
	return command.NoError()
}

func mkSingleCmdCLI() cli.CLI {
	// make a new CLI passing the version string and a command.Cmd
	// while it's possible to add flags to the CLI, they are ignored; only the
	// Cmd's flags are presented to the user.
	return cli.New("1.0.2", Cmd())
}

// This example shows how to create a single-command CLI
func Example_singleCommand() {
	// this would be your main() function

	// run the Main function, which calls os.Exit with the appropriate exit status
	mkSingleCmdCLI().Main()
}

// Add the following to your tests to validate that there are no collisions
// between command flags and that help text can be generated without error:

// package main

// import (
// 	"testing"

// 	"github.com/turbinelabs/test/assert"
// )

// func TestCLI(t *testing.T) {
// 	assert.Nil(t, mkCLI().Validate())
// }
