// The cli package provides a simple library for creating command-line
// applications with multiple sub-commands. It supports both global and
// per-subcommand flags, and automatically generates help and version
// sub-commands.
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/turbinelabs/tbn/cli/app"
	"github.com/turbinelabs/tbn/cli/command"
	"github.com/turbinelabs/tbn/cli/flags"
)

// A CLI represents a command-line application
type CLI struct {
	Flags flag.FlagSet // The global FlagSet for the CLI

	usage       app.Usage
	commands    []*command.Cmd
	versionFlag bool
	helpFlag    bool

	// stub out some method calls for easier unit testing
	fillFlagsFromEnv func(*flag.FlagSet)
	osArgs           func() []string
	stderr           func(string)
	exit             func(int)
}

// NewCLI produces a CLI for the given app.App and command.Cmds.
func NewCLI(app app.App, commands ...*command.Cmd) *CLI {
	cli := new(CLI)
	cli.usage = app.Usage()

	addVersionFlag(&cli.Flags, &cli.versionFlag)
	addHelpFlag(&cli.Flags, &cli.helpFlag)

	help := command.Help(
		cli.command,
		cli.globalUsage,
		cli.commandUsage,
	)
	version := command.Version(app.Name, app.Version)
	cli.commands = append(append(commands, help), version)

	cli.fillFlagsFromEnv = func(fs *flag.FlagSet) {
		flags.FillFromEnv(app.Name, fs)
	}

	cli.osArgs = func() []string { return os.Args }
	cli.stderr = func(msg string) { fmt.Fprint(os.Stderr, msg) }
	cli.exit = os.Exit

	return cli
}

// Main serves as the main() function for the CLI. It will parse
// the command-line arguments and flags, call the appropriate sub-command,
// and return exit status and output error messages as appropriate.
func (cli *CLI) Main() {
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

	cli.exit(int(cmdErr.Code))
}

func addVersionFlag(fs *flag.FlagSet, flag *bool) {
	fs.BoolVar(flag, "version", false, command.VersionSummary)
}

func addHelpFlag(fs *flag.FlagSet, flag *bool) {
	fs.BoolVar(flag, "help", false, command.HelpSummary)
}

func (cli *CLI) mainOrCmdErr() command.CmdErr {
	err := cli.Flags.Parse(cli.osArgs()[1:])
	if err != nil {
		return command.BadInput(err)
	}

	// fill unset flags from env
	cli.fillFlagsFromEnv(&cli.Flags)
	args := cli.Flags.Args()

	if len(args) < 1 {
		// <app> -help
		if (cli.helpFlag) {
			cli.globalUsage()
			return command.NoError()
		}
		return command.BadInput("no command specified")
	}

	// <app> -version <ignored>
	if cli.versionFlag {
		args[0] = "version"
	}

	cmdHelpFlag := false

	// determine which Cmd should be run, parse args
	if cmd := cli.command(args[0]); cmd != nil {
		// only add help flag if not already present
		if cmd.Flags.Lookup("help") == nil {
			addHelpFlag(&cmd.Flags, &cmdHelpFlag)
		}
		// <app> -help <command>
		if (cli.helpFlag) {
			cli.commandUsage(cmd)
			return command.NoError()
		}
		// parse flags
		if err := cmd.Flags.Parse(args[1:]); err != nil {
			return cmd.BadInput(err)
		}
		// fill unset flags from env
		cli.fillFlagsFromEnv(&cmd.Flags)
		// <app> <command> -help
		if cmdHelpFlag {
			cli.commandUsage(cmd)
			return command.NoError()
		}
		// run the command
		return cmd.Run()
	}

	// <app> -help <unknown command>
	if (cli.helpFlag) {
		cli.globalUsage()
		return command.NoError()
	}

	// if we got this far, the specified command is bogus, no
	// global help flag has been set, but we still need to check for
	// the command-level help flag
	var badCmdFlags flag.FlagSet
	addHelpFlag(&badCmdFlags, &cmdHelpFlag)
	// ignore errors, since we're only trying to get the help flag
	badCmdFlags.Parse(args[1:])

	// <app> <unknown command> -help
	if (cmdHelpFlag) {
		cli.globalUsage()
		return command.NoError()
	}

	return command.BadInputf("unknown command: %q", args[0])
}

func (cli *CLI) command(name string) *command.Cmd {
	for _, c := range cli.commands {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (cli *CLI) globalUsage() {
	cli.usage.Global(cli.commands, &cli.Flags)
}

func (cli *CLI) commandUsage(cmd *command.Cmd) {
	cli.usage.Command(cmd)
}
