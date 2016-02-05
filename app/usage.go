package app

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/turbinelabs/tbn/cli/command"
	"github.com/turbinelabs/tbn/cli/flags"
)

// Usage provides methods to output global and command-specific usage of an App.
// No specific output format or destination is specified by the interface.
type Usage interface {
	// Global outputs the global usage for an App, based on the provided
	// command.Cmds and flag.FlagSet
	Global(cmds []*command.Cmd, fs *flag.FlagSet)
	// Command outputs the usage for the given command.Cmd
	Command(cmd *command.Cmd)
}

// The default implementation of Usage for this App, prints tab-formatted
// output to STDOUT.
type usageT struct {
	app                  App
	globalUsageTemplate  *template.Template
	commandUsageTemplate *template.Template
	tabWriter            *tabwriter.Writer
}

const (
	globalUsageTemplateStr = `
NAME:
{{printf "\t%s - %s" .Executable .Description}}

USAGE:
{{printf "\t%s" .Executable}} [global options] <command> [command options] [arguments...]

VERSION:
{{printf "\t%s" .Version}}

COMMANDS:{{range .Commands}}
{{printf "\t%s\t%s" .Name .Summary}}{{end}}

GLOBAL OPTIONS:{{range .Flags}}
{{printOption .Name .DefValue .Usage}}{{end}}

Global options can also be configured via upper-case environment variables prefixed with "{{.EXECUTABLE}}"
For example, "--some_flag" --> "{{.EXECUTABLE}}_SOME_FLAG"

Run "{{.Executable}} help <command>" for more details on a specific command.
`
	commandUsageTemplateStr = `
NAME:
{{printf "\t%s - %s" .Cmd.Name .Cmd.Summary}}

USAGE:
{{printf "\t%s %s %s" .Executable .Cmd.Name .Cmd.Usage}}

DESCRIPTION:
{{range $line := descToLines .Cmd.Description}}{{printf "\t%s" $line}}
{{end}}
{{if .CmdFlags}}OPTIONS:{{range .CmdFlags}}
{{printOption .Name .DefValue .Usage}}{{end}}

{{end}}For help on global options run "{{.Executable}} help"
`
)

func newUsage(a App, wr io.Writer) Usage {
	tabWriter := new(tabwriter.Writer)
	tabWriter.Init(wr, 0, 8, 1, '\t', 0)

	templFuncs := template.FuncMap{
		"descToLines": func(s string) []string {
			// trim leading/trailing whitespace and split into slice of lines
			return strings.Split(strings.Trim(s, "\n\t "), "\n")
		},
		"printOption": func(name, defvalue, usage string) string {
			prefix := "--"
			if len(name) == 1 {
				prefix = "-"
			}
			return fmt.Sprintf("\t%s%s=%s\t%s", prefix, name, defvalue, usage)
		},
	}

	globalUsageTemplate := template.Must(
		template.New("global_usage").Funcs(templFuncs).Parse(globalUsageTemplateStr[1:]))

	commandUsageTemplate := template.Must(
		template.New("command_usage").Funcs(templFuncs).Parse(commandUsageTemplateStr[1:]))

	return usageT{a, globalUsageTemplate, commandUsageTemplate, tabWriter}
}

func (u usageT) Global(cmds []*command.Cmd, fs *flag.FlagSet) {
	u.globalUsageTemplate.Execute(u.tabWriter, struct {
		Executable  string
		EXECUTABLE  string
		Commands    []*command.Cmd
		Flags       []*flag.Flag
		Description string
		Version     string
	}{
		u.app.Name,
		strings.ToUpper(u.app.Name),
		cmds,
		flags.Enumerate(fs),
		u.app.Description,
		u.app.Version,
	})
	u.tabWriter.Flush()
}

func (u usageT) Command(cmd *command.Cmd) {
	u.commandUsageTemplate.Execute(u.tabWriter, struct {
		Executable string
		Cmd        *command.Cmd
		CmdFlags   []*flag.Flag
	}{
		u.app.Name,
		cmd,
		flags.Enumerate(&cmd.Flags),
	})
	u.tabWriter.Flush()
}
