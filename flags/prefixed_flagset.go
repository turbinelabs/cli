package flags

//go:generate ./prefixed_flagset_gen.sh bool time.Duration float64 int int64 string uint uint64

import (
	"flag"
	"strings"
)

type PrefixedFlagSet struct {
	*flag.FlagSet

	prefix     string
	descriptor string
}

func NewPrefixedFlagSet(fs *flag.FlagSet, prefix, descriptor string) *PrefixedFlagSet {
	if prefix != "" && !strings.HasSuffix(prefix, ".") {
		prefix = prefix + "."
	}

	return &PrefixedFlagSet{
		FlagSet:    fs,
		prefix:     prefix,
		descriptor: descriptor,
	}
}

func (f *PrefixedFlagSet) mkUsage(usage string) string {
	return strings.Replace(usage, "{{NAME}}", f.descriptor, -1)
}

func (f *PrefixedFlagSet) Var(value flag.Value, name string, usage string) {
	f.FlagSet.Var(value, f.prefix+name, f.mkUsage(usage))
}
