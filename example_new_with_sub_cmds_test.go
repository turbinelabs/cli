/*
Copyright 2017 Turbine Labs, Inc.

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
	"strings"

	"github.com/turbinelabs/cli"
	"github.com/turbinelabs/cli/command"
)

// The typical pattern is to provide a public CmdXYZ() func for each
// sub-command you wish to provide. This function should initialize the
// command.Cmd, the command.Runner, and flags.
func CmdSplit() *command.Cmd {
	// typically the command.Runner is initialized only with internally-defined
	// state; all necessary external state should be provided via flags. One can
	// inline the initializaton of the command.Runner in the command.Cmd
	// initialization if no flags are necessary, but it's often convenient to
	// have a typed reference
	runner := &splitRunner{}

	cmd := &command.Cmd{
		Name:        "split",
		Summary:     "split strings",
		Usage:       "[OPTIONS] <string>",
		Description: "split strings using the specified delimiter",
		Runner:      runner,
	}

	// The flag.FlagSet is a member of the command.Cmd, and the flag
	// value is a member of the command.Runner.
	cmd.Flags.StringVar(&runner.delim, "delim", ",", "The delimiter on which to split the string")

	return cmd
}

// The private command.Runner implementation should contain any state needed
// to execute the command. The values should be initialized via flags declared
// in the CmdXYZ() function.
type splitRunner struct {
	delim string
}

// Run does the actual work, based on state provided by flags, and the
// args remaining after the flags have been parsed.
func (f *splitRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	// argument validation should occur at the top of the function, and
	// errors should be reported via the cmd.BadInput or cmd.BadInputf methods
	if len(args) < 1 {
		return cmd.BadInput("missing \"string\" argument.")
	}
	str := args[0]
	if globalFlags.verbose {
		fmt.Printf("Splitting \"%s\"\n", str)
	}
	split := strings.Split(str, f.delim)
	for i, term := range split {
		if globalFlags.verbose {
			fmt.Printf("[%d] ", i)
		}
		fmt.Println(term)
	}

	// In this case, there was no error. Errors should be returned via the
	// cmd.Error or cmd.Errorf methods.
	return command.NoError()
}

// A second command
func CmdJoin() *command.Cmd {
	runner := &joinRunner{}

	cmd := &command.Cmd{
		Name:        "join",
		Summary:     "join strings",
		Usage:       "[OPTIONS] <string>...",
		Description: "join strings using the specified delimiter",
		Runner:      runner,
	}

	cmd.Flags.StringVar(&runner.delim, "delim", ",", "The delimiter with which to join the strings")

	return cmd
}

// a second Runner
type joinRunner struct {
	delim string
}

func (f *joinRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	if globalFlags.verbose {
		fmt.Printf("Joining \"%v\"\n", args)
	}
	joined := strings.Join(args, f.delim)
	fmt.Println(joined)

	return command.NoError()
}

// while not manditory, keeping globally-configured flags in a single struct
// makes it obvious where they came from at access time.
type globalFlagsT struct {
	verbose bool
}

var globalFlags = globalFlagsT{}

func mkSubCmdCLI() cli.CLI {
	// make a new CLI passing the description and version and one or more sub commands
	c := cli.NewWithSubCmds(
		"an example CLI for simple string operations",
		"1.2.3",
		CmdSplit(),
		CmdJoin(),
	)

	// Global flags can be used to modify global state
	c.Flags().BoolVar(&globalFlags.verbose, "verbose", false, "Produce verbose output")

	return c
}

// This example shows how to create a CLI with multiple sub-commands
func Example_subCommands() {
	// this would be your main() function

	// run the Main function, which calls os.Exit with the appropriate exit status
	mkSubCmdCLI().Main()
}

// Add the following to your tests to validate that there are no collisions
// between command flags:

// package main

// import (
// 	"testing"

// 	"github.com/turbinelabs/test/assert"
// )

// func TestCLI(t *testing.T) {
// 	assert.Nil(t, mkCLI().Validate())
// }
