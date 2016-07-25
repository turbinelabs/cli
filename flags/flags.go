// The flags package provides convenience methods for dealing with flag.FlagSets
package flags

import (
	"flag"
	"strings"
)

const requiredPrefix = "[REQUIRED] "

// Required prefixes its argument with "[REQUIRED] " which, in addition from
// documenting for the users of the command.Cmd on which the flag is declared
// that the argument is required, will also cause it to be checked when the
// Cmd's Run method is invoked.
func Required(usage string) string {
	return requiredPrefix + usage
}

// IsRequired checks the usage string of the given Flag to see if it is
// prefixed with "[REQUIRED] ".
func IsRequired(f *flag.Flag) bool {
	return strings.HasPrefix(f.Usage, requiredPrefix)
}

// AllRequired produces a slice of the names of all flags for which the Usage
// string is prefxied with "[REQUIRED] ".
func AllRequired(fs *flag.FlagSet) []string {
	result := []string{}
	fs.VisitAll(func(f *flag.Flag) {
		if IsRequired(f) {
			result = append(result, f.Name)
		}
	})
	return result
}

// MissingRequired produces a slice of the names of all flags for which the
// Usage string is prefixed with "[REQUIRED] " but no value has been set.
func MissingRequired(fs *flag.FlagSet) []string {
	seen := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		seen[f.Name] = true
	})

	result := []string{}
	fs.VisitAll(func(f *flag.Flag) {
		if !seen[f.Name] && IsRequired(f) {
			result = append(result, f.Name)
		}
	})

	return result
}

// Enumerate returns a slice containing all Flags in the Flagset
func Enumerate(flagset *flag.FlagSet) []*flag.Flag {
	if flagset == nil {
		return []*flag.Flag{}
	}
	flags := make([]*flag.Flag, 0)
	flagset.VisitAll(func(f *flag.Flag) {
		flags = append(flags, f)
	})
	return flags
}
