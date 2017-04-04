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

package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/turbinelabs/cli/app"
	"github.com/turbinelabs/cli/command"
	tbnflag "github.com/turbinelabs/nonstdlib/flag"
	"github.com/turbinelabs/nonstdlib/flag/usage"
	"github.com/turbinelabs/nonstdlib/log/console"
	tbnos "github.com/turbinelabs/nonstdlib/os"
)

const HelpSummary = "Show a list of commands or help for one command"
const VersionSummary = "Print the version and exit"

type ValidationFlag int

const (
	// Skips Validating that global and subcommand help text can
	// be generated.
	ValidateSkipHelpText ValidationFlag = iota
)

// A CLI represents a command-line application
type CLI interface {
	// Flags returns a pointer to the global flags for the CLI
	Flags() *flag.FlagSet
	// Set the flags
	SetFlags(*flag.FlagSet)

	// Main serves as the main() function for the CLI. It will parse
	// the command-line arguments and flags, call the appropriate sub-command,
	// and return exit status and output error messages as appropriate.
	Main()

	// Validate can be used to make sure the CLI is well-defined from within
	// unit tests. In particular it will validate that no two flags exist with
	// the same environment key. As a last-ditch effort, Validate will be called
	// at the start of Main. ValidationFlag values may be passed to alter the
	// level of validation performed.
	Validate(...ValidationFlag) error

	// Returns the CLI version data.
	Version() app.Version
}

type cli struct {
	flags flag.FlagSet // The global FlagSet for the CLI

	commands    []*command.Cmd
	name        string
	app         app.App
	usage       app.Usage
	version     app.Version
	versionFlag bool
	helpFlag    bool

	flagsFromEnv    tbnflag.FromEnv
	cmdFlagsFromEnv map[string]tbnflag.FromEnv

	os tbnos.OS
}

// New produces a CLI for the given command.Cmd
func New(version string, command *command.Cmd) CLI {
	app := app.App{
		Name:          path.Base(os.Args[0]),
		Description:   command.Description,
		VersionString: version,
		HasSubCmds:    false,
	}
	return mkNew(app, command)
}

// NewWithSubCmds produces a CLI for the given app.App and with subcommands
// for the given command.Cmds.
func NewWithSubCmds(
	description string,
	version string,
	command1 *command.Cmd,
	command2 *command.Cmd,
	commandsN ...*command.Cmd,
) CLI {
	app := app.App{
		Name:          path.Base(os.Args[0]),
		Description:   description,
		VersionString: version,
		HasSubCmds:    true,
	}
	commands := []*command.Cmd{command1}
	commands = append(commands, command2)
	commands = append(commands, commandsN...)
	return mkNew(app, commands...)
}

func mkNew(app app.App, commands ...*command.Cmd) CLI {
	c := &cli{
		commands: commands,
		app:      app,
		name:     app.Name,
		usage:    app.Usage(),
		version:  app.Version(),

		os: tbnos.New(),
	}

	c.flagsFromEnv = tbnflag.NewFromEnv(&c.flags, app.Name)

	if len(commands) == 1 {
		c.cmdFlagsFromEnv = map[string]tbnflag.FromEnv{
			app.Name: tbnflag.NewFromEnv(&commands[0].Flags, app.Name),
		}
	} else {
		c.cmdFlagsFromEnv = map[string]tbnflag.FromEnv{}
		for _, cmd := range commands {
			c.cmdFlagsFromEnv[cmd.Name] = tbnflag.NewFromEnv(&cmd.Flags, app.Name, cmd.Name)
		}
	}

	return c
}

func validateFlagIsSet(vflags []ValidationFlag, vflag ValidationFlag) bool {
	for _, f := range vflags {
		if f == vflag {
			return true
		}
	}

	return false
}

func (cli *cli) Validate(vflags ...ValidationFlag) error {
	seen := map[string]string{}
	collisions := map[string][]string{}

	// add top-level flags
	for _, f := range tbnflag.Enumerate(&cli.flags) {
		seen[tbnflag.EnvKey(cli.name, f.Name)] = fmt.Sprintf("%s -%s", cli.name, f.Name)
	}

	// add cmd-level flags
	for _, cmd := range cli.commands {
		for _, f := range tbnflag.Enumerate(&cmd.Flags) {
			envKey := tbnflag.EnvKey(cli.name, cmd.Name, f.Name)
			cmdWithArg := fmt.Sprintf("%s %s -%s", cli.name, cmd.Name, f.Name)
			if seen[envKey] != "" {
				// if we've seen it before, it's a problem
				if len(collisions[envKey]) == 0 {
					// add the first seen
					collisions[envKey] = []string{seen[envKey]}
				}
				// add this one
				collisions[envKey] = append(collisions[envKey], cmdWithArg)
			} else {
				// otherwise, mark as seen
				seen[envKey] = cmdWithArg
			}
		}
	}

	if len(collisions) > 0 {
		msg := "possible environment key collisions:\n"
		for k, vs := range collisions {
			msg += fmt.Sprintf("  %s: \"%s\"\n", k, strings.Join(vs, `", "`))
		}
		return errors.New(msg)
	}

	if !validateFlagIsSet(vflags, ValidateSkipHelpText) {
		if err := cli.validateHelpText(); err != nil {
			return err
		}
	}

	return nil
}

func (cli *cli) validateHelpText() error {
	usage := cli.app.RedirectedUsage(bytes.NewBufferString(""))
	errs := []string{}

	try := func(u func()) {
		defer func() {
			if e := recover(); e != nil {
				switch v := e.(type) {
				case error:
					errs = append(errs, v.Error())
				default:
					errs = append(errs, fmt.Sprintf("%v", e))
				}
			}
		}()

		u()
	}

	try(func() {
		usage.Global(cli.commands, cli.flagsFromEnv)
	})
	for _, cmd := range cli.commands {
		try(func() {
			usage.Command(cmd, cli.flagsFromEnv, cli.commandFlagsFromEnv(cmd))
		})
	}

	if len(errs) > 0 {
		msg := "error(s) generating help text:\n  "
		msg += strings.Join(errs, "\n  ")
		return errors.New(msg + "\n")
	}

	return nil
}

func (cli *cli) Version() app.Version {
	return cli.version
}

func (cli *cli) Main() {
	if err := cli.Validate(ValidateSkipHelpText); err != nil {
		cli.stderr(fmt.Sprintf("%s\n\n", err))
		cli.os.Exit(2)
	}

	cmdErr := cli.mainOrCmdErr()

	if cmdErr.IsError() {
		cli.stderr(fmt.Sprintf("%s\n\n", cmdErr.Message))
	}

	if cmdErr.Code == command.CmdErrCodeBadInput {
		if cmd := cmdErr.Cmd; cmd != nil {
			cli.commandUsage(cmd)
		} else {
			cli.globalUsage()
		}
	}

	cli.os.Exit(int(cmdErr.Code))
}

func (cli *cli) Flags() *flag.FlagSet {
	return &cli.flags
}

func (cli *cli) SetFlags(fs *flag.FlagSet) {
	cli.flags = *fs
}

func (cli *cli) mainOrCmdErr() command.CmdErr {
	// if we have a single command, it is implicitly the first argument
	if len(cli.commands) == 1 {
		return cli.cmdOrCmdErr(cli.commands[0], cli.os.Args(), []string{})
	}

	args, err := cli.parseGlobalFlags()
	if err != nil {
		return command.BadInput(err)
	}

	if cli.versionFlag {
		// <app> version [ignored]
		// <app> -version [ignored]
		// <app> -v [ignored]
		fmt.Println(cli.version.Describe())
		return command.NoError()
	}

	checkDeprecated(&cli.flags, "global ")

	missingErrs := checkRequired(&cli.flags, []string{}, "global ")

	if len(args) < 1 {
		// <app> help
		// <app> -help
		// <app> -h
		if cli.helpFlag {
			cli.globalUsage()
			return command.NoError()
		}

		errs := append([]string{"no command specified"}, missingErrs...)
		return command.BadInput(strings.Join(errs, "\n"))
	}

	// determine which Cmd should be run, parse args
	if cmd := cli.command(args[0]); cmd != nil {
		return cli.cmdOrCmdErr(cmd, args, missingErrs)
	}

	return cli.handleBadCmd(args, missingErrs)
}

func (cli *cli) parseGlobalFlags() ([]string, error) {
	addVersionFlagIfMissing(&cli.flags, &cli.versionFlag)
	addHelpFlagIfMissing(&cli.flags, &cli.helpFlag)

	// parse flags
	if err := quietParse(&cli.flags, cli.os.Args()[1:]); err != nil {
		return nil, err
	}

	// fill unset flags from env
	cli.flagsFromEnv.Fill()
	args := cli.flags.Args()

	// treat help as -help
	if len(args) > 0 && args[0] == "help" {
		cli.helpFlag = true
		args = args[1:]
	}

	// treat version as -version
	if len(args) > 0 && args[0] == "version" {
		cli.versionFlag = true
		args = args[1:]
	}

	return args, nil
}

func (cli *cli) cmdOrCmdErr(cmd *command.Cmd, args []string, missingErrs []string) command.CmdErr {
	// only add help flag if not already present
	var cmdHelpFlag bool
	addHelpFlagIfMissing(&cmd.Flags, &cmdHelpFlag)

	// only add version flag if not already present
	var cmdVersionFlag bool
	addVersionFlagIfMissing(&cmd.Flags, &cmdVersionFlag)

	// parse flags
	if err := quietParse(&cmd.Flags, args[1:]); err != nil {
		return cmd.BadInput(err)
	}

	// fill unset flags from env
	cli.commandFlagsFromEnv(cmd).Fill()

	// <app> <command> -help
	// <app> <command> -h
	// <app> help <command>
	// <app> -help <command>
	// <app> -h <command>
	if cmdHelpFlag || cli.helpFlag {
		cli.commandUsage(cmd)
		return command.NoError()
	}

	// <app> <command> -version
	// <app> <command> -v
	if cmdVersionFlag {
		fmt.Println(cli.version.Describe())
		return command.NoError()
	}

	checkDeprecated(&cmd.Flags, "")

	missingErrs = checkRequired(&cmd.Flags, missingErrs, "")
	if len(missingErrs) > 0 {
		return cmd.BadInputf("\n  %s", strings.Join(missingErrs, "\n  "))
	}

	// run the command
	return cmd.Run()
}

func (cli *cli) handleBadCmd(args []string, validationErrs []string) command.CmdErr {
	// <app> help <unknown command>
	// <app> -help <unknown command>
	// <app> -h <unknown command>
	if cli.helpFlag {
		cli.globalUsage()
		return command.NoError()
	}

	// if we got this far, the specified command is bogus, no
	// global help flag has been set, but we still need to check for
	// the command-level help flag
	var badCmdFlags flag.FlagSet
	var badCmdHelpFlag bool
	addHelpFlagIfMissing(&badCmdFlags, &badCmdHelpFlag)
	// ignore errors, since we're only trying to get the help flag
	quietParse(&badCmdFlags, args[1:])

	// <app> <unknown command> -help
	// <app> <unknown command> -h
	if badCmdHelpFlag {
		cli.globalUsage()
		return command.NoError()
	}

	errs := append([]string{fmt.Sprintf("unknown command: %q", args[0])}, validationErrs...)
	return command.BadInput(strings.Join(errs, "\n"))
}

func (cli *cli) command(name string) *command.Cmd {
	for _, c := range cli.commands {
		if strings.ToLower(c.Name) == strings.ToLower(name) {
			return c
		}
	}
	return nil
}

func (cli *cli) globalUsage() {
	cli.usage.Global(cli.commands, cli.flagsFromEnv)
}

func (cli *cli) commandUsage(cmd *command.Cmd) {
	cli.usage.Command(cmd, cli.flagsFromEnv, cli.commandFlagsFromEnv(cmd))
}

func (cli *cli) commandFlagsFromEnv(cmd *command.Cmd) tbnflag.FromEnv {
	if len(cli.commands) == 1 {
		return cli.cmdFlagsFromEnv[cli.name]
	} else {
		return cli.cmdFlagsFromEnv[cmd.Name]
	}
}

func (cli *cli) stderr(msg string) {
	fmt.Fprint(cli.os.Stderr(), msg)
}

func checkRequired(fs *flag.FlagSet, errStrs []string, prefix string) []string {
	for _, name := range usage.MissingRequired(fs) {
		errStrs = append(errStrs, fmt.Sprintf("--%s is a required %sflag", name, prefix))
	}
	return errStrs
}

func checkDeprecated(fs *flag.FlagSet, prefix string) {
	for _, name := range usage.DeprecatedAndSet(fs) {
		console.Error().Printf("%sflag --%s is deprecated", prefix, name)
	}
}

func quietParse(fs *flag.FlagSet, args []string) error {
	fs.SetOutput(ioutil.Discard)
	return fs.Parse(args)
}

func addVersionFlagIfMissing(fs *flag.FlagSet, flag *bool) {
	if fs.Lookup("version") == nil {
		fs.BoolVar(flag, "version", false, VersionSummary)
	}
	if fs.Lookup("v") == nil {
		fs.BoolVar(flag, "v", false, VersionSummary)
	}
}

func addHelpFlagIfMissing(fs *flag.FlagSet, flag *bool) {
	if fs.Lookup("help") == nil {
		fs.BoolVar(flag, "help", false, HelpSummary)
	}
	if fs.Lookup("h") == nil {
		fs.BoolVar(flag, "h", false, HelpSummary)
	}
}
