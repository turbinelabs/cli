// The flags package provides convenience methods for dealing with flag.FlagSets
package flags

import (
	"flag"
	"os"
	"strings"
)

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

func fillFromEnv(prefix string, fs *flag.FlagSet, getenv func(string) string) {
	alreadySet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		alreadySet[f.Name] = true
	})
	fs.VisitAll(func(f *flag.Flag) {
		if !alreadySet[f.Name] {
			key := strings.ToUpper(strings.Replace(prefix+"_"+f.Name, "-", "_", -1))
			val := getenv(key)
			if val != "" {
				fs.Set(f.Name, val)
			}
		}
	})
}
