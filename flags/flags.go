// The flags package provides convenience methods for dealing with flag.FlagSets
package flags

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

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

// FillFromEnv parses all registered flags in the given flagset,
// and if they are not already set it attempts to set their values from
// environment variables. Environment variables take the name of the flag but
// are UPPERCASE, have the given prefix, and any dashes are replaced by
// underscores. For example:
// 	some-flag -> PREFIX_SOME_FLAG
func FillFromEnv(prefix string, fs *flag.FlagSet) {
	fillFromEnv(prefix, fs, os.Getenv)
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

func fillFromEnv(prefix string, fs *flag.FlagSet, getenv func(string) string) {
	alreadySet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		alreadySet[f.Name] = true
	})
	fs.VisitAll(func(f *flag.Flag) {
		if !alreadySet[f.Name] {
			key := EnvKey(prefix, f.Name)
			val := getenv(key)
			if val != "" {
				fs.Set(f.Name, val)
			}
		}
	})
}
