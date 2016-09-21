package flags

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Integer time values larger than this are assumed to be in
// milliseconds since the Unix epoch. Treated as millisecons since the
// epoch, this timestamp represents 1973-01-11T16:26:24Z. Treated as
// seconds, this timestamp represents 5000-01-01:00:00:00Z.
const TimestampConversionFencepost = int64(95617584000)

// Timestamp conforms to the flag.Value and flag.Getter interfaces. It
// can be used to populate a timestamp from a command line
// argument. It accepts the following inputs:
//
// - timestamps in time.RFC3339Nano format (fractional seconds
//   optional)
//
// - integer seconds since the Unix epoch (see
//   TimestampConversionFencepost)
//
// - integer milliseconds since the Unix epoch (see
//   TimestampConversionFencepost)
//
// - "now" (case insensitive)
type Timestamp struct {
	Value time.Time
}

var _ flag.Getter = &Timestamp{}

func NewTimestamp(defaultTime time.Time) Timestamp {
	return Timestamp{Value: defaultTime}
}

func (t *Timestamp) Set(value string) error {
	if strings.ToLower(value) == "now" {
		t.Value = time.Now()
		return nil
	} else if ticks, err := strconv.ParseInt(value, 10, 64); err == nil {
		if ticks >= TimestampConversionFencepost {
			t.Value = time.Unix(ticks/1000, (ticks%1000)*int64(time.Millisecond)).UTC()
		} else {
			t.Value = time.Unix(ticks, 0).UTC()
		}
		return nil
	} else if ts, err := time.Parse(time.RFC3339Nano, value); err == nil {
		t.Value = ts
		return nil
	}

	return fmt.Errorf(
		"cannot parse '%s': expecting seconds or milliseconds since the Unix epoch or RFC3339 format (fractional seconds optional)",
		value,
	)
}

func (t *Timestamp) Get() interface{} {
	return t.Value
}

func (t *Timestamp) String() string {
	return t.Value.Format(time.RFC3339Nano)
}
