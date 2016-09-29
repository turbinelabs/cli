package flags

import (
	"flag"
	"fmt"
	"strings"

	"github.com/turbinelabs/arrays/indexof"
)

// Choice conforms to the flag.Value and flag.Getter interfaces, and
// can be used populate a slice of strings from a flag.Flag.
type Choice struct {
	// Populated from the command line.
	Choice *string

	// All possible values allowed to appear in Choice.
	AllowedValues []string
}

var _ flag.Getter = &Choice{}

// NewChoice produces a Choice with a set of allowed values.
func NewChoice(allowedValues ...string) Choice {
	return Choice{AllowedValues: allowedValues}
}

func (cv Choice) WithDefault(value string) Choice {
	cv.Set(value)
	return cv
}

func (cv *Choice) String() string {
	if cv.Choice != nil {
		return *cv.Choice
	}
	return ""
}

func (cv *Choice) Set(value string) error {
	if indexof.String(cv.AllowedValues, value) == indexof.NotFound {
		return fmt.Errorf(
			"invalid flag value: %s, must be one of %s",
			value,
			strings.Join(cv.AllowedValues, ", "),
		)
	}

	cv.Choice = &value
	return nil
}

func (cv *Choice) Get() interface{} {
	return cv.Choice
}
