package terminal

import (
	"strings"

	tbnos "github.com/turbinelabs/nonstdlib/os"
)

// Ask formats and prints a prompt (which should be a yes or no
// question) on stdout, and then waits for a line of input from
// stdin. The string " [y/N]: " is automatically appended to the
// prompt. After trimming whitespace and converting to lowercase, if
// the response is "y" or "yes", the function returns true. An error
// may occur reading stdin.
func Ask(os tbnos.OS, yesNoQ string, args ...interface{}) (bool, error) {
	response, err := Prompt(os, yesNoQ+" [y/N]: ", args...)
	if err != nil {
		return false, err
	}

	trimmed := strings.ToLower(strings.TrimSpace(response))
	return trimmed == "y" || trimmed == "yes", nil
}
