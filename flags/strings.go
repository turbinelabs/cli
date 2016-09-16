package flags

import (
	"flag"
	"fmt"
	"strings"

	"github.com/turbinelabs/arrays/indexof"
)

// Strings conforms to the flag.Value and flag.Getter interfaces, and
// can be used populate a slice of strings from a flag.Flag.
type Strings struct {
	// Populated from the command line.
	Strings []string

	// All possible values allowed to appear in Strings. An empty
	// slice means any value is allowed in Strings.
	AllowedValues []string

	// Delimiter used to parse the string from the command line.
	Delimiter string
}

var _ flag.Getter = &Strings{}

// NewStrings produces a Strings with the default delimiter (",").
func NewStrings() Strings {
	return Strings{Delimiter: ","}
}

// NewStrings produces a Strings with a set of allowed values and the
// default delimiter (",").
func NewStringsWithConstraint(allowedValues []string) Strings {
	return Strings{AllowedValues: allowedValues, Delimiter: ","}
}

func (ssv *Strings) String() string {
	return strings.Join(ssv.Strings, ssv.Delimiter)
}

func (ssv *Strings) Set(value string) error {
	parts := strings.Split(value, ssv.Delimiter)

	disallowed := []string{}

	i := 0
	for i < len(parts) {
		parts[i] = strings.TrimSpace(parts[i])
		if parts[i] == "" {
			if i+1 > len(parts) {
				parts = parts[0:i]
			} else {
				parts = append(parts[0:i], parts[i+1:]...)
			}
		} else {
			if len(ssv.AllowedValues) > 0 {
				if indexof.String(ssv.AllowedValues, parts[i]) == indexof.NotFound {
					disallowed = append(disallowed, parts[i])
				}
			}
			i++
		}

	}

	if len(disallowed) > 0 {
		return fmt.Errorf(
			"invalid flag value(s): %s",
			strings.Join(disallowed, ssv.Delimiter+" "),
		)
	}

	ssv.Strings = parts

	return nil
}

func (ssv *Strings) Get() interface{} {
	return ssv.Strings
}
