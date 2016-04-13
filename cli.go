// The cli package provides a simple library for creating command-line
// applications with multiple sub-commands. It supports both global and
// per-subcommand flags, and automatically generates help and version
// sub-commands.
package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/turbinelabs/cli/app"
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/cli/flags"
)

const HelpSummary = "Show a list of commands or help for one command"
const VersionSummary = "Print the version and exit"

// A CLI represents a command-line application
type CLI struct {
	Flags flag.FlagSet // The global FlagSet for the CLI

	commands    []*command.Cmd
	usage       app.Usage
	version     app.Version
	versionFlag bool
	helpFlag    bool

	// stub out some method calls for easier unit testing
	fillFlagsFromEnv func(*flag.FlagSet)
	osArgs           func() []string
	stderr           func(string)
	exit             func(int)
}

// New produces a CLI for the given command.Cmd
func New(version string, command *command.Cmd) *CLI {
	app := app.App{
		Name:          os.Args[0],
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
) *CLI {
	app := app.App{
		Name:          os.Args[0],
		Description:   description,
		VersionString: version,
		HasSubCmds:    true,
	}
	commands := []*command.Cmd{command1}
	commands = append(commands, command2)
	commands = append(commands, commandsN...)
	return mkNew(app, commands...)
}

func mkNew(app app.App, commands ...*command.Cmd) *CLI {
	return &CLI{
		commands: commands,
		usage:    app.Usage(),
		version:  app.Version(),

		fillFlagsFromEnv: func(fs *flag.FlagSet) {
			flags.FillFromEnv(app.Name, fs)
		},

		osArgs: func() []string { return os.Args },
		stderr: func(msg string) { fmt.Fprint(os.Stderr, msg) },
		exit:   os.Exit,
	}
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

func (cli *CLI) mainOrCmdErr() command.CmdErr {
	// if we have a single command, it is implicitly the first argument
	if len(cli.commands) == 1 {
		return cli.cmdOrCmdErr(cli.commands[0], cli.osArgs())
	}

	args, err := cli.parseGlobalFlags()
	if err != nil {
		return command.BadInput(err)
	}

	if cli.versionFlag {
		// <app> version [ignored]
		// <app> -version [ignored]
		// <app> -v [ignored]
		cli.version.Print()
		return command.NoError()
	}

	if len(args) < 1 {
		// <app> help
		// <app> -help
		// <app> -h
		if cli.helpFlag {
			cli.globalUsage()
			return command.NoError()
		}

		return command.BadInput("no command specified")
	}

	// determine which Cmd should be run, parse args
	if cmd := cli.command(args[0]); cmd != nil {
		return cli.cmdOrCmdErr(cmd, args)
	}

	return cli.handleBadCmd(args)
}

func (cli *CLI) parseGlobalFlags() ([]string, error) {
	addVersionFlagIfMissing(&cli.Flags, &cli.versionFlag)
	addHelpFlagIfMissing(&cli.Flags, &cli.helpFlag)

	// parse flags
	if err := quietParse(&cli.Flags, cli.osArgs()[1:]); err != nil {
		return nil, err
	}

	// fill unset flags from env
	cli.fillFlagsFromEnv(&cli.Flags)
	args := cli.Flags.Args()

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

func (cli *CLI) cmdOrCmdErr(cmd *command.Cmd, args []string) command.CmdErr {
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
	cli.fillFlagsFromEnv(&cmd.Flags)

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
		cli.version.Print()
		return command.NoError()
	}

	// run the command
	return cmd.Run()
}

func (cli *CLI) handleBadCmd(args []string) command.CmdErr {
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
