// The command package provides an abstraction for a command-line application
// sub-command, a means to execute code when that sub-command is invoked, a
// means to report success/failure status of said code, and generic
// implementations of help and version sub-commands.
package command

import (
	"flag"
	"fmt"
)

// A Runner represents the executable code associated with a Cmd. Typically
// the struct implementing Runner will include whatever state is needed,
// configured by the Flags in the associated Cmd.
type Runner interface {
	// Execute code associated with a Cmd with the given arguments,
	// return exit status. The Cmd is provided here to avoid a circular
	// dependency between the Runner and the Cmd.
	Run(cmd *Cmd, args []string) CmdErr
}

// A Cmd represents a named sub-command for a command-line application.
type Cmd struct {
	Name        string       // Name of the Command and the string to use to invoke it
	Summary     string       // One-sentence summary of what the Command does
	Usage       string       // Usage options/arguments
	Description string       // Detailed description of command
	Flags       flag.FlagSet // Set of flags associated with this Cmd, which typically configure the Runner
	Runner      Runner       // The code to run when this Cmd is invoked
}

// Run invokes the Runner associated with this Cmd, passing the args remaining
// after flags are parsed out.
func (c *Cmd) Run() CmdErr {
	if c.Runner == nil {
		return c.Error("No Runner specified")
	}
	return c.Runner.Run(c, c.Flags.Args())
}

// BadInputf produces a Cmd-scoped CmdErr with an exit code of 2, based on the
// given format string and args, which are passed to fmt.Sprintf.
func (c *Cmd) BadInputf(format string, args ...interface{}) CmdErr {
	return c.BadInput(fmt.Sprintf(format, args...))
}

// Errorf produces a Cmd-scoped CmdErr with an exit code of 1, based on the
// given format string and args, which are passed to fmt.Sprintf.
func (c *Cmd) Errorf(format string, args ...interface{}) CmdErr {
	return c.Error(fmt.Sprintf(format, args...))
}

// BadInput produces a Cmd-scoped CmdErr with an exit code of 2, based on the
// given args, which are passed to fmt.Sprint.
func (c *Cmd) BadInput(args ...interface{}) CmdErr {
	return CmdErr{c, CmdErrCodeBadInput, fmt.Sprintf("%s: %s", c.Name, fmt.Sprint(args...))}
}

// Error produces a Cmd-scoped CmdErr with an exit code of 1, based on the
// given args, which are passed to fmt.Sprint.
func (c *Cmd) Error(args ...interface{}) CmdErr {
	return CmdErr{c, CmdErrCodeError, fmt.Sprintf("%s: %s", c.Name, fmt.Sprint(args...))}
}

// CmdErrCode is the exit code for the application
type CmdErrCode uint32

const (
	CmdErrCodeNoError  CmdErrCode = 0 // No Error
	CmdErrCodeError               = 1 // Generic Error
	CmdErrCodeBadInput            = 2 // Bad Input Error
)

// CmdErr represents the exit status of a Cmd.
type CmdErr struct {
	Cmd     *Cmd       // The Cmd that produced the exit status. Can be nil for global errors
	Code    CmdErrCode // The exit code
	Message string     // Additional information if the Code is non-zero
}

// IsError returns true if the exit code is non-zero
func (err CmdErr) IsError() bool {
	return err.Code != CmdErrCodeNoError
}

var cmdErrNoErr = CmdErr{nil, CmdErrCodeNoError, ""}

// BadInputf produces an unscoped CmdErr with an exit code of 2, based on the
// given format string and args, which are passed to fmt.Sprintf.
func BadInputf(format string, args ...interface{}) CmdErr {
	return BadInput(fmt.Sprintf(format, args...))
}

// Errorf produces an uscoped CmdErr with an exit code of 1, based on the
// given format string and args, which are passed to fmt.Sprintf.
func Errorf(format string, args ...interface{}) CmdErr {
	return Error(fmt.Sprintf(format, args...))
}

// BadInput produces an unscoped CmdErr with an exit code of 2, based on the
// given args, which are passed to fmt.Sprint.
func BadInput(args ...interface{}) CmdErr {
	return CmdErr{nil, CmdErrCodeBadInput, fmt.Sprint(args...)}
}

// Error produces an uncoped CmdErr with an exit code of 1, based on the
// given args, which are passed to fmt.Sprint.
func Error(args ...interface{}) CmdErr {
	return CmdErr{nil, CmdErrCodeError, fmt.Sprint(args...)}
}

// NoError produces an uncoped CmdErr with an exit code of 0
func NoError() CmdErr {
	return cmdErrNoErr
}
