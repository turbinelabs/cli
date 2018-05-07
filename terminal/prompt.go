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
	"fmt"
	"io"

	tbnos "github.com/turbinelabs/nonstdlib/os"
)

// Prompt formats a prompt, prints it on stdout, and then waits for a
// line of input from stdin. The trailing CRLF or LF is not
// returned. If the response ends with an EOF, the characters read up
// to EOF are returned with the EOF error. Prompt uses unbuffered
// input to avoid consuming bytes intended for subsequent consumers of
// stdin.
func Prompt(os tbnos.OS, prompt string, args ...interface{}) (string, error) {
	fmt.Fprintf(os.Stdout(), prompt, args...)

	buffer := make([]byte, 128)
	in := os.Stdin()
	n := 0
	eof := false
	for {
		nr, err := in.Read(buffer[n : n+1])
		if err == io.EOF {
			n--
			eof = true
			break
		}

		if err != nil {
			return "", err
		}

		if nr > 0 {
			if buffer[n] == '\n' {
				break
			}

			n++
			if n >= len(buffer) {
				expanded := make([]byte, len(buffer)*2)
				copy(expanded, buffer)
				buffer = expanded
			}
		}
	}

	if n >= 0 && buffer[n] == '\n' {
		n--
	}

	for n >= 0 && buffer[n] == '\r' {
		n--
	}

	var err error
	if eof {
		err = io.EOF
	}
	if n >= 0 {
		return string(buffer[:n+1]), err
	}

	return "", err
}
