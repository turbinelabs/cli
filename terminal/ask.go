/*
Copyright 2018 Turbine Labs, Inc.

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
