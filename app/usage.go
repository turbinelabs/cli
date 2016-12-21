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

package app

//go:generate mockgen -source $GOFILE -destination mock_$GOFILE -package $GOPACKAGE

import (
	"bytes"
	"flag"
	"fmt"
	"go/doc"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"
	"unicode"

	tty "github.com/mattn/go-isatty"

	"github.com/turbinelabs/cli/command"
	tbnflag "github.com/turbinelabs/nonstdlib/flag"
)

const widthFromTerm = -1

// Usage provides methods to output global and command-specific usage of an App.
// No specific output format or destination is specified by the interface.
type Usage interface {
	// Global outputs the global usage for an App, based on the provided
	// command.Cmds and flag.FlagSet
	Global(cmds []*command.Cmd, flagsFromEnv tbnflag.FromEnv)
	// Command outputs the usage for the given command.Cmd, including global and
	// command flags.
	Command(cmd *command.Cmd, globalFlagsFromEnv tbnflag.FromEnv, cmdFlagsFromEnv tbnflag.FromEnv)
}

// The default implementation of Usage for this App, prints tab-formatted
// output to STDOUT.
type usageT struct {
	app                  App
	globalUsageTemplate  *template.Template
	commandUsageTemplate *template.Template
	tabWriter            *tabwriter.Writer
	width                int
}

const (
	rightIndent     = 4 // The right-side padding for cleaned text
	preformatIndent = 4 // The extra indent for preformatted text

	globalUsageTemplateStr = `
{{bold "NAME"}}
{{cleanf 4 "%s - %s" .Executable .Description}}
{{bold "USAGE"}}
{{cleanf 4 "%s [GLOBAL OPTIONS] <command> [COMMAND OPTIONS] [arguments...]" .Executable}}
{{bold "VERSION"}}
{{clean 4 .Version}}
{{bold "COMMANDS"}}{{range .Commands}}
{{cmd .Name .Summary}}{{end}}
{{bold "GLOBAL OPTIONS"}}{{range .GlobalFlags.AllFlags}}{{option .}}{{end}}
{{optionsText "Global options" .GlobalFlags.Prefix .GlobalFlags.Filled}}
{{- cmdHelp .Executable}}
`
	commandUsageTemplateStr = `
{{bold "NAME"}}
{{if .HasSubCmds}}{{cleanf 4 "%s - %s" .Cmd.Name .Cmd.Summary}}
{{else}}{{cleanf 4 "%s - %s" .Executable .Cmd.Summary}}
{{end}}{{bold "USAGE"}}
{{if .HasSubCmds}}{{cleanf 4 "%s [GLOBAL OPTIONS] %s %s" .Executable .Cmd.Name .Cmd.Usage}}
{{else}}{{cleanf 4 "%s %s" .Executable .Cmd.Usage}}
{{end}}{{bold "VERSION"}}
{{clean 4 .Version}}
{{bold "DESCRIPTION"}}
{{clean 4 .Cmd.Description}}
{{if .HasSubCmds}}{{bold "GLOBAL OPTIONS"}}{{range .GlobalFlags.AllFlags}}{{option .}}{{end}}
{{optionsText "Global options" .GlobalFlags.Prefix .GlobalFlags.Filled}}{{end -}}
{{if .CmdFlags}}{{bold "OPTIONS"}}{{range .CmdFlags.AllFlags}}{{option .}}{{end}}
{{optionsText "Options" .CmdFlags.Prefix .CmdFlags.Filled}}{{end}}`
)

func notGraphic(r rune) bool {
	return !unicode.IsGraphic(r)
}

var consoleIsInteractive = true

func ul(s string) string {
	if consoleIsInteractive {
		return "\033[4m" + s + "\033[0m"
	}
	return s
}

func bold(s string) string {
	if consoleIsInteractive {
		return "\033[1m" + s + "\033[0m"
	}
	return s
}

func termWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	bytes, err := cmd.Output()
	if err != nil {
		return 80
	}
	pair := strings.SplitN(string(bytes), " ", 2)
	w, err := strconv.Atoi(strings.TrimSpace(pair[1]))
	if err != nil {
		return 80
	}
	return w
}

func (u usageT) cleanf(indent int, format string, args ...interface{}) string {
	return u.clean(indent, fmt.Sprintf(format, args...))
}

// replace whitespace with a single space, indent and wrap
func (u usageT) clean(indent int, s string) string {
	templFuncs := template.FuncMap{
		"bold": bold,
		"ul":   ul,
	}

	var buf bytes.Buffer
	cleanTmpl := template.Must(template.New("clean").Funcs(templFuncs).Parse(s))
	cleanTmpl.Execute(&buf, nil)

	var res bytes.Buffer
	doc.ToText(
		&res,
		buf.String(),
		strings.Repeat(" ", indent),
		strings.Repeat(" ", indent+preformatIndent),
		u.width-indent-rightIndent,
	)
	return res.String()
}

// print a flag, including defaults
func (u usageT) option(f *flag.Flag) string {
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
	if defValue != "" {
		if typeName == "string" || strings.IndexFunc(defValue, notGraphic) != -1 {
			defValue = fmt.Sprintf("%q", defValue)
		}
		defValue = u.cleanf(12, "(default: %s)", defValue)
	}
	result := ""
	nameLen := len(prefix + f.Name + eq + typeName)
	fullName := "    " + prefix + ul(f.Name) + eq + typeName
	//     --name=type (default: x)
	if nameLen < 7 {
		result += fullName
		if defValue != "" {
			result += fmt.Sprintf("%s%s", strings.Repeat(" ", 8-nameLen), defValue[12:])
		}
	} else {
		//     --longerName=type
		//            (default: x)
		result += u.clean(4, fullName)
		if defValue != "" {
			result += defValue
		}
	}
	result += u.clean(12, usage)
	return "\n" + result
}

func (u usageT) cmd(name, desc string) string {
	cleanDesc := u.clean(12, desc)
	if len(name) < 7 && len(cleanDesc) < u.width {
		cleanDesc = u.clean(0, desc)
		return fmt.Sprintf(
			"    %s%s%s",
			ul(name),
			strings.Repeat(" ", 8-len(name)),
			cleanDesc,
		)
	} else {
		return fmt.Sprintf("    %s\n%s", ul(name), cleanDesc)
	}
}

func (u usageT) optionsText(prefix string, envKey string, flagsFromEnv map[string]string) string {
	format := `%s can also be configured via upper-case, underscore-delimeted environment variables
prefixed with "%s". For example, "--some-flag" becomes "%sSOME_FLAG". Command-line flags take
precedence over environment variables.`

	if len(flagsFromEnv) > 0 {
		format += "\n\nOptions currently configured from the Environment:"
	}
	result := u.cleanf(4, format, prefix, envKey, envKey) + "\n"

	lines := make([]string, len(flagsFromEnv))
	for key, value := range flagsFromEnv {
		lines = append(lines, u.cleanf(8, "%s=%s", key, value))
	}
	sort.Strings(lines)
	if len(lines) > 0 {
		lines = append(lines, "\n")
	}

	return result + strings.Join(lines, "")
}

func (u usageT) globalHelp(name string) string {
	return u.cleanf(0, `For global options run "%s help".`, name)
}

func (u usageT) cmdHelp(name string) string {
	return u.cleanf(0, `Run "%s help <command>" for more details on a specific command.`, name)
}

func newUsage(a App, wr io.Writer, width int, forceInteractive bool) Usage {
	if width == widthFromTerm {
		width = termWidth()
	}
	tabWriter := new(tabwriter.Writer)
	tabWriter.Init(wr, 0, 8, 1, '\t', 0)

	if !forceInteractive {
		consoleIsInteractive = tty.IsTerminal(os.Stdout.Fd())
	}

	u := usageT{app: a, tabWriter: tabWriter, width: width}

	templFuncs := template.FuncMap{
		"bold":        bold,
		"ul":          ul,
		"clean":       u.clean,
		"cmd":         u.cmd,
		"cleanf":      u.cleanf,
		"option":      u.option,
		"optionsText": u.optionsText,
		"globalHelp":  u.globalHelp,
		"cmdHelp":     u.cmdHelp,
	}

	u.globalUsageTemplate = template.Must(
		template.New("global_usage").Funcs(templFuncs).Parse(globalUsageTemplateStr[1:]))

	u.commandUsageTemplate = template.Must(
		template.New("command_usage").Funcs(templFuncs).Parse(commandUsageTemplateStr[1:]))

	return u
}

func (u usageT) Global(cmds []*command.Cmd, flagsFromEnv tbnflag.FromEnv) {
	u.globalUsageTemplate.Execute(u.tabWriter, struct {
		Executable  string
		Commands    []*command.Cmd
		GlobalFlags tbnflag.FromEnv
		Description string
		Version     string
	}{
		u.app.Name,
		cmds,
		flagsFromEnv,
		u.app.Description,
		u.app.VersionString,
	})
	u.tabWriter.Flush()
}

func (u usageT) Command(
	cmd *command.Cmd,
	globalFlagsFromEnv tbnflag.FromEnv,
	cmdFlagsFromEnv tbnflag.FromEnv,
) {
	u.commandUsageTemplate.Execute(u.tabWriter, struct {
		Executable  string
		HasSubCmds  bool
		Cmd         *command.Cmd
		GlobalFlags tbnflag.FromEnv
		CmdFlags    tbnflag.FromEnv
		Version     string
	}{
		u.app.Name,
		u.app.HasSubCmds,
		cmd,
		globalFlagsFromEnv,
		cmdFlagsFromEnv,
		u.app.VersionString,
	})
	u.tabWriter.Flush()
}
