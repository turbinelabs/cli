package flags_test

import (
	"flag"
	"fmt"

	"github.com/turbinelabs/cli/flags"
)

func ExampleNewStrings() {
	var strings flags.Strings // typically a field in a struct
	strings = flags.NewStrings()

	flagset := flag.NewFlagSet("example", flag.PanicOnError)
	flagset.Var(
		&strings,
		"x",
		"Flag help.",
	)

	flagset.Parse([]string{"-x=a,b,c"})

	for _, selected := range strings.Strings {
		fmt.Println(selected)
	}
	// Output:
	// a
	// b
	// c
}

func ExampleNewStringsWithConstraint() {
	var strings flags.Strings // typically a field in a struct
	strings = flags.NewStringsWithConstraint([]string{"choice1", "option2", "possibility3"})

	flagset := flag.NewFlagSet("example", flag.PanicOnError)
	flagset.Var(
		&strings,
		"x",
		"Flag help.",
	)

	flagset.Parse([]string{"-x=choice1,possibility3"})

	for _, selected := range strings.Strings {
		fmt.Println(selected)
	}
	// Output:
	// choice1
	// possibility3
}

func ExampleStrings_withDelimiter() {
	var strings flags.Strings // typically a field in a struct
	strings = flags.Strings{Delimiter: ";"}

	flagset := flag.NewFlagSet("example", flag.PanicOnError)
	flagset.Var(
		&strings,
		"x",
		"Flag help.",
	)

	flagset.Parse([]string{"-x=one;two"})

	for _, selected := range strings.Strings {
		fmt.Println(selected)
	}
	// Output:
	// one
	// two
}

func ExampleNewChoice() {
	var choice flags.Choice // typically a field in a struct
	choice = flags.NewChoice([]string{"a", "b", "c"})

	flagset := flag.NewFlagSet("example", flag.PanicOnError)
	flagset.Var(
		&choice,
		"x",
		"Flag help.",
	)

	flagset.Parse([]string{"-x=c"})

	fmt.Println(choice.String())
	// Output:
	// c
}
