package app

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"text/template"
	"unicode"

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

GLOBAL OPTIONS:{{range .Flags}}{{printOption .}}{{end}}

Global options can also be configured via upper-case environment variables prefixed with "{{.EnvPrefix}}"
For example, "--some_flag" --> "{{.EnvPrefix}}SOME_FLAG"

Run "{{.Executable}} help <command>" for more details on a specific command.
`
	commandUsageTemplateStr = `
NAME:
{{if .HasSubCmds}}{{printf "\t%s - %s" .Cmd.Name .Cmd.Summary}}
{{else}}{{printf "\t%s - %s" .Executable .Cmd.Summary}}
{{end}}
USAGE:
{{if .HasSubCmds}}{{printf "\t%s %s %s" .Executable .Cmd.Name .Cmd.Usage}}
{{else}}{{printf "\t%s %s" .Executable .Cmd.Usage}}
{{end}}
VERSION:
{{printf "\t%s" .Version}}

DESCRIPTION:
{{range $line := descToLines .Cmd.Description}}{{printf "\t%s" $line}}
{{end}}
{{if .CmdFlags}}OPTIONS:{{range .CmdFlags}}{{printOption .}}{{end}}{{end}}

Options can also be configured via upper-case environment variables prefixed with "{{.EnvPrefix}}"
For example, "--some_flag" --> "{{.EnvPrefix}}SOME_FLAG"
{{if .HasSubCmds}}
For help on global options run "{{.Executable}} help"
{{end}}`
)

func notGraphic(r rune) bool {
	return !unicode.IsGraphic(r)
}

func newUsage(a App, wr io.Writer) Usage {
	tabWriter := new(tabwriter.Writer)
	tabWriter.Init(wr, 0, 8, 1, '\t', 0)

	templFuncs := template.FuncMap{
		"descToLines": func(s string) []string {
			// trim leading/trailing whitespace and split into slice of lines
			return strings.Split(strings.Trim(s, "\n\t "), "\n")
		},
		"printOption": func(f *flag.Flag) string {
			typeName, usage := flag.UnquoteUsage(f)
			if f.Name == "h" || f.Name == "v" {
				return ""
			}
			prefix := "--"
			if len(f.Name) == 1 {
				prefix = "-"
			}
			// flags doesn't export types, so we have to cheat
			eq := "="
			if typeName == "" {
				eq = ""
			}
			defValue := f.DefValue
			if typeName == "string" || strings.IndexFunc(defValue, notGraphic) != -1 {
				defValue = fmt.Sprintf("%q", f.DefValue)
			}

			if defValue != "" {
				defValue = fmt.Sprintf("(default: %s)", defValue)
			}

			//--things=int (default: 5)
			return fmt.Sprintf(
				"\n\t%s%s%s%s\t%s\t%s",
				prefix,
				f.Name,
				eq,
				typeName,
				defValue,
				usage,
			)
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
		EnvPrefix   string
		Commands    []*command.Cmd
		Flags       []*flag.Flag
		Description string
		Version     string
	}{
		u.app.Name,
		flags.EnvKey(u.app.Name, ""),
		cmds,
		flags.Enumerate(fs),
		u.app.Description,
		u.app.VersionString,
	})
	u.tabWriter.Flush()
}

func (u usageT) Command(cmd *command.Cmd) {
	u.commandUsageTemplate.Execute(u.tabWriter, struct {
		Executable string
		EnvPrefix  string
		HasSubCmds bool
		Cmd        *command.Cmd
		CmdFlags   []*flag.Flag
		Version    string
	}{
		u.app.Name,
		flags.EnvKey(u.app.Name, ""),
		u.app.HasSubCmds,
		cmd,
		flags.Enumerate(&cmd.Flags),
		u.app.VersionString,
	})
	u.tabWriter.Flush()
}
