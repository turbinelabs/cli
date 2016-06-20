// The flags package provides convenience methods for dealing with flag.FlagSets
package flags

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const requiredPrefix = "[REQUIRED] "

// Required prefixes its argument with "[REQUIRED] " which, in addition from
// documenting for the users of the `command.Cmd` on which the flag is declared
// that the argument is required, will also cause it to be checked when the
// `Cmd`'s `Run` method is invoked.
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

var notAlphaNum, _ = regexp.Compile("[^A-Za-z0-9_]+")

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

// FillFromEnv parses all registered flags in the given FlagSet,
// and if they are not already set it attempts to set their values from
// environment variables. Environment variables take the name of the flag but
// are UPPERCASE, have the given prefix, and any dashes are replaced by
// underscores. For example:
// 	some-flag -> PREFIX_SOME_FLAG
//
// the provided map[string]string is also populated with the keys and values
// added to the FlagSet.
func FillFromEnv(prefix string, fs *flag.FlagSet, filledFromEnv map[string]string) {
	fillFromEnv(prefix, fs, filledFromEnv, os.Getenv)
}

// EnvKey produces a namespaced environment variable key, concatonates a prefix
// and key with an infix underscore, replacing all non-alphanumeric,
// non-underscope characters with underscores, and upper-casing the entire
// string
func EnvKey(prefix, key string) string {
	return strings.ToUpper(fmt.Sprintf(
		"%s_%s",
		notAlphaNum.ReplaceAllString(prefix, "_"),
		notAlphaNum.ReplaceAllString(key, "_"),
	))
}

func fillFromEnv(
	prefix string,
	fs *flag.FlagSet,
	filledFromEnv map[string]string,
	getenv func(string) string,
) {
	alreadySet := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		alreadySet[f.Name] = true
	})
	fs.VisitAll(func(f *flag.Flag) {
		if !alreadySet[f.Name] {
			key := EnvKey(prefix, f.Name)
			val := getenv(key)
			if val != "" {
				fs.Set(f.Name, val)
				filledFromEnv[key] = val
			}
		}
	})
}
