package command

const HelpSummary = "Show a list of commands or help for one command"

// Help produces a generic help Cmd, based on the supplied function arguments:
// 	command     func(string) *Cmd // looks up the Cmd with the given name, or nil
// 	globalUsage func()            // outputs global application usage
// 	cmdUsage    func(*Cmd)        // outputs Cmd-specific usage
func Help(command func(string) *Cmd, globalUsage func(), cmdUsage func(*Cmd)) *Cmd {
	return &Cmd{
		Name:        "help",
		Summary:     HelpSummary,
		Usage:       "[COMMAND]",
		Description: "Show a list of commands or detailed help for one command",
		Runner:      &helpRunner{command, globalUsage, cmdUsage},
	}
}

type helpRunner struct {
	command     func(string) *Cmd
	globalUsage func()
	cmdUsage    func(*Cmd)
}

func (h *helpRunner) Run(cmd *Cmd, args []string) CmdErr {
	if len(args) < 1 {
		h.globalUsage()
		return NoError()
	}

	if cmd := h.command(args[0]); cmd != nil {
		h.cmdUsage(cmd)
		return NoError()
	}

	return cmd.BadInputf("unknown command: %q", args[0])
}
