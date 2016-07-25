package flags

//go:generate mockgen -source $GOFILE -destination mock_$GOFILE -package $GOPACKAGE

import (
	"flag"
	"regexp"
	"strings"

	tbnos "github.com/turbinelabs/os"
)

var (
	notAlphaNum         = regexp.MustCompile("[^A-Za-z0-9_]+")
	multipleUnderscores = regexp.MustCompile("_+")
)

// NewFromEnv produces a FromEnv, using the provided FlagSet and scopes.
// The scopes are used to produce the environment key prefix, by uppercasing,
// replacing non-alphanumeric+underscore characters with underscores, and
// concatenating with underscores.
//
// For example:
//  {"foo-foo", "bar.bar", "baz"} -> "FOO_FOO_BAR_BAR_BAZ"
func NewFromEnv(fs *flag.FlagSet, scopes ...string) FromEnv {
	return fromEnv{
		prefix:        EnvKey(scopes...),
		fs:            fs,
		os:            tbnos.New(),
		filledFromEnv: map[string]string{},
	}
}

// FromEnv supports operations on a FlagSet based on environment variables.
// In particular, FromEnv allows one to fill a FlagSet from the environment
// and then inspect the results.
type FromEnv interface {
	// Prefix returns the environment key prefix, eg "SOME_PREFIX_"
	Prefix() string

	// Fill parses all registered flags in the FlagSet, and if they are not already
	// set it attempts to set their values from environment variables. Environment
	// variables take the name of the flag but are UPPERCASE, have the given prefix,
	// and non-alphanumeric+underscore chars are replaced by underscores.
	//
	// For example:
	//  some-flag -> SOME_PREFIX_SOME_FLAG
	//
	// the provided map[string]string is also populated with the keys and values
	// added to the FlagSet.
	Fill()

	// Filled returns a map of the environment keys and values for flags currently
	// filled from the environment.
	Filled() map[string]string

	// AllFlags returns a slice containing all Flags in the underlying Flagset
	AllFlags() []*flag.Flag
}

type fromEnv struct {
	prefix        string
	fs            *flag.FlagSet
	os            tbnos.OS
	filledFromEnv map[string]string
}

func (fe fromEnv) Prefix() string {
	return EnvKey(fe.prefix, "")
}

func (fe fromEnv) Fill() {
	alreadySet := map[string]bool{}
	fe.fs.Visit(func(f *flag.Flag) {
		alreadySet[f.Name] = true
	})
	fe.fs.VisitAll(func(f *flag.Flag) {
		if !alreadySet[f.Name] {
			key := EnvKey(fe.prefix, f.Name)
			val := fe.os.Getenv(key)
			if val != "" {
				fe.fs.Set(f.Name, val)
				fe.filledFromEnv[key] = val
			}
		}
	})
}

func (fe fromEnv) Filled() map[string]string {
	return fe.filledFromEnv
}

func (fe fromEnv) AllKeys() []string {
	keys := []string{}
	fe.fs.VisitAll(func(f *flag.Flag) {
		keys = append(keys, EnvKey(fe.prefix, f.Name))
	})
	return keys
}

func (fe fromEnv) AllFlags() []*flag.Flag {
	return Enumerate(fe.fs)
}

// EnvKey produces a namespaced environment variable key, concatenates a prefix
// and key with an infix underscore, replacing all non-alphanumeric,
// non-underscore characters with underscores, and upper-casing the entire
// string
func EnvKey(parts ...string) string {
	for i, part := range parts {
		parts[i] = notAlphaNum.ReplaceAllString(part, "_")
	}
	joined := strings.ToUpper(strings.Join(parts, "_"))
	return multipleUnderscores.ReplaceAllLiteralString(joined, "_")
}
