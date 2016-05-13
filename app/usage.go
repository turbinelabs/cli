package app

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

	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/cli/flags"
)

// Usage provides methods to output global and command-specific usage of an App.
// No specific output format or destination is specified by the interface.
type Usage interface {
	// Global outputs the global usage for an App, based on the provided
	// command.Cmds and flag.FlagSet
	Global(cmds []*command.Cmd, fs *flag.FlagSet, flagsFromEnv map[string]string)
	// Command outputs the usage for the given command.Cmd
	Command(cmd *command.Cmd, flagsFromEnv map[string]string)
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
	globalUsageTemplateStr = `
{{bold "NAME"}}
{{cleanf 4 "%s - %s" .Executable .Description}}
{{bold "USAGE"}}
{{cleanf 4 "%s [global options] <command> [command options] [arguments...]" .Executable}}
{{bold "VERSION"}}
{{clean 4 .Version}}
{{bold "COMMANDS"}}{{range .Commands}}
{{cmd .Name .Summary}}{{end}}
{{bold "GLOBAL OPTIONS"}}{{range .Flags}}{{option .}}{{end}}
{{optionsText "Global options" .EnvPrefix .FlagsFromEnv}}
{{cmdHelp .Executable}}`
	commandUsageTemplateStr = `
{{bold "NAME"}}
{{if .HasSubCmds}}{{cleanf 4 "%s - %s" .Cmd.Name .Cmd.Summary}}
{{else}}{{cleanf 4 "%s - %s" .Executable .Cmd.Summary}}
{{end}}{{bold "USAGE"}}
{{if .HasSubCmds}}{{cleanf 4 "%s %s %s" .Executable .Cmd.Name .Cmd.Usage}}
{{else}}{{cleanf 4 "%s %s" .Executable .Cmd.Usage}}
{{end}}{{bold "VERSION"}}
{{clean 4 .Version}}
{{bold "DESCRIPTION"}}
{{cleanWithPre 4 .Cmd.Description}}
{{if .CmdFlags}}{{bold "OPTIONS"}}{{range .CmdFlags}}{{option .}}{{end}}
{{optionsText "Options" .EnvPrefix .FlagsFromEnv}}{{end}}
{{- if .HasSubCmds}}
{{globalHelp .Executable}}{{end}}`
)

func notGraphic(r rune) bool {
	return !unicode.IsGraphic(r)
}

func ul(s string) string {
	return "\033[4m" + s + "\033[0m"
}

func bold(s string) string {
	return "\033[1m" + s + "\033[0m"
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
	clean := strings.Join(strings.Fields(s), " ")
	var b bytes.Buffer
	doc.ToText(&b, clean, strings.Repeat(" ", indent), "", u.width-indent)
	return b.String()
}

// lines with leading 4-space indent are treated as pre-formatted text,
// all other lines are processed as with `clean`.
func (u usageT) cleanWithPre(indent int, s string) string {
	lines := strings.Split(s, "\n")
	var clean string
	previousWasPreformatted := false
	for _, line := range lines {
		if strings.HasPrefix(line, "    ") {
			if !previousWasPreformatted {
				clean += "\n"
			}
			clean += line + "\n"
			previousWasPreformatted = true
		} else {
			if previousWasPreformatted {
				previousWasPreformatted = false
			} else {
				clean += " "
			}
			clean += strings.Join(strings.Fields(line), " ")
		}
	}

	var b bytes.Buffer
	doc.ToText(
		&b,
		clean,
		strings.Repeat(" ", indent),
		strings.Repeat(" ", indent*2),
		u.width-indent)
	return b.String()
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
	if len(name) < 7 && len(cleanDesc) < termWidth() {
		return fmt.Sprintf("    %s%s%s", ul(name), strings.Repeat(" ", 8-len(name)), cleanDesc[12:])
	} else {
		return fmt.Sprintf("    %s\n%s", ul(name), cleanDesc)
	}
}

func (u usageT) optionsText(prefix, envKey string, flagsFromEnv map[string]string) string {
	format := `%s can also be configured via upper-case environment variables prefixed with "%s"
For example, "--some-flag" --> "%sSOME_FLAG". Command-line flags take precidence over environment variables.`
	if len(flagsFromEnv) > 0 {
		format += " Options currently configured from the Environment:"
	}
	result := u.cleanf(0, format, prefix, envKey, envKey) + "\n"

	lines := make([]string, len(flagsFromEnv))
	for key, value := range flagsFromEnv {
		lines = append(lines, u.cleanf(4, "%s=%s", key, value))
	}
	sort.Strings(lines)

	return result + strings.Join(lines, "")
}

func (u usageT) globalHelp(name string) string {
	return u.cleanf(0, `For global options run "%s help".`, name)
}

func (u usageT) cmdHelp(name string) string {
	return u.cleanf(0, `Run "%s help <command>" for more details on a specific command.`, name)
}

func newUsage(a App, wr io.Writer, width int) Usage {
	if width == -1 {
		width = termWidth()
	}
	tabWriter := new(tabwriter.Writer)
	tabWriter.Init(wr, 0, 8, 1, '\t', 0)

	u := usageT{app: a, tabWriter: tabWriter, width: width}

	templFuncs := template.FuncMap{
		"bold":         bold,
		"ul":           ul,
		"clean":        u.clean,
		"cleanWithPre": u.cleanWithPre,
		"cmd":          u.cmd,
		"cleanf":       u.cleanf,
		"option":       u.option,
		"optionsText":  u.optionsText,
		"globalHelp":   u.globalHelp,
		"cmdHelp":      u.cmdHelp,
	}

	u.globalUsageTemplate = template.Must(
		template.New("global_usage").Funcs(templFuncs).Parse(globalUsageTemplateStr[1:]))

	u.commandUsageTemplate = template.Must(
		template.New("command_usage").Funcs(templFuncs).Parse(commandUsageTemplateStr[1:]))

	return u
}

func (u usageT) Global(cmds []*command.Cmd, fs *flag.FlagSet, flagsFromEnv map[string]string) {
	u.globalUsageTemplate.Execute(u.tabWriter, struct {
		Executable   string
		EnvPrefix    string
		Commands     []*command.Cmd
		Flags        []*flag.Flag
		FlagsFromEnv map[string]string
		Description  string
		Version      string
	}{
		u.app.Name,
		flags.EnvKey(u.app.Name, ""),
		cmds,
		flags.Enumerate(fs),
		flagsFromEnv,
		u.app.Description,
		u.app.VersionString,
	})
	u.tabWriter.Flush()
}

func (u usageT) Command(cmd *command.Cmd, flagsFromEnv map[string]string) {
	u.commandUsageTemplate.Execute(u.tabWriter, struct {
		Executable   string
		EnvPrefix    string
		HasSubCmds   bool
		Cmd          *command.Cmd
		CmdFlags     []*flag.Flag
		FlagsFromEnv map[string]string
		Version      string
	}{
		u.app.Name,
		flags.EnvKey(u.app.Name, ""),
		u.app.HasSubCmds,
		cmd,
		flags.Enumerate(&cmd.Flags),
		flagsFromEnv,
		u.app.VersionString,
	})
	u.tabWriter.Flush()
}
