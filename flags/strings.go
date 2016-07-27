package flags

import "strings"

// Strings conformes to the flag.Value and flag.Getter interfaces, and
// can be used populate a slice of strings from a flag.Flag.
type Strings struct {
	Strings   []string
	Delimiter string
}

// NewStrings produces a Strings with the default delimiter (",").
func NewStrings() Strings {
	return Strings{Delimiter: ","}
}

func (ssv *Strings) String() string {
	return strings.Join(ssv.Strings, ssv.Delimiter)
}

func (ssv *Strings) Set(value string) error {
	ssv.Strings = strings.Split(value, ssv.Delimiter)
	return nil
}

func (ssv *Strings) Get() interface{} {
	return ssv.Strings
}
