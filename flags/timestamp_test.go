package flags

import (
	"fmt"
	"testing"
	"time"

	"github.com/turbinelabs/test/assert"
)

const (
	inSeconds      = "1474406514"
	inMilliseconds = "1474406514572"
	inRFC3339      = "2016-09-20T21:21:54Z"
	inRFC3339Nano  = "2016-09-20T21:21:54.572Z"
)

var (
	tsSeconds = time.Unix(1474406514, 0).UTC()
	tsNanos   = time.Unix(1474406514, 572000000).UTC()
)

func TestTimestampSetHandlesNow(t *testing.T) {
	ts := Timestamp{tsNanos}

	before := time.Now()
	assert.Nil(t, ts.Set("Now"))
	after := time.Now()

	assert.True(t, before.Before(ts.Value) || before.Equal(ts.Value))
	assert.True(t, after.After(ts.Value) || after.Equal(ts.Value))
}

func TestTimestampSetHandlesSecondsSinceEpoch(t *testing.T) {
	ts := Timestamp{time.Now()}

	assert.Nil(t, ts.Set(inSeconds))
	assert.Equal(t, ts.Value, tsSeconds)
}

func TestTimestampSetHandlesMillisecondsSinceEpoch(t *testing.T) {
	ts := Timestamp{time.Now()}

	assert.Nil(t, ts.Set(inMilliseconds))
	assert.Equal(t, ts.Value, tsNanos)
}

func TestTimestampFencepost(t *testing.T) {
	testcases := [][]string{
		{fmt.Sprintf("%d", TimestampConversionFencepost), "1973-01-11T16:26:24Z"},
		{fmt.Sprintf("%d", TimestampConversionFencepost+1), "1973-01-11T16:26:24.001Z"},
		{fmt.Sprintf("%d", TimestampConversionFencepost-1), "4999-12-31T23:59:59Z"},
	}

	for _, tc := range testcases {
		value := tc[0]
		expected := tc[1]

		ts := Timestamp{}
		assert.Nil(t, ts.Set(value))
		assert.Equal(t, ts.Value.Format(time.RFC3339Nano), expected)
	}
}

func TestTimestampSetHandlesRFC3339(t *testing.T) {
	ts := Timestamp{time.Now()}

	assert.Nil(t, ts.Set(inRFC3339))
	assert.Equal(t, ts.Value, tsSeconds)
}

func TestTimestampSetHandlesRFC3339Nano(t *testing.T) {
	ts := Timestamp{time.Now()}

	assert.Nil(t, ts.Set(inRFC3339Nano))
	assert.Equal(t, ts.Value, tsNanos)
}

func TestTimestampSetReportsFormatError(t *testing.T) {
	ts := Timestamp{time.Now()}

	assert.ErrorContains(t, ts.Set("Tue, 20 September 2016 21:21:54 UTC"), "cannot parse")
}

func TestTimestampGetReturnsValue(t *testing.T) {
	now := time.Now()
	ts := Timestamp{now}

	assert.DeepEqual(t, ts.Get(), now)
}

func TestTimestampStringUsesRFC3339Nano(t *testing.T) {
	ts := Timestamp{tsNanos}

	assert.Equal(t, ts.String(), inRFC3339Nano)
}
