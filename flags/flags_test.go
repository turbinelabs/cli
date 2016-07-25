package flags

import (
	"flag"
	"testing"

	"github.com/turbinelabs/test/assert"
)

func testFlags() (*flag.FlagSet, *string, *string) {
	var fs flag.FlagSet
	fooFlag := fs.String("foo-baz", "", "do the foo")
	barFlag := fs.String("bar", "", Required("harty har to the bar"))
	return &fs, fooFlag, barFlag
}

func TestRequired(t *testing.T) {
	assert.Equal(t, Required("foo"), "[REQUIRED] foo")
}

func TestIsRequired(t *testing.T) {
	assert.True(t, IsRequired(&flag.Flag{Usage: Required("foo")}))
	assert.False(t, IsRequired(&flag.Flag{}))
}

func TestMissingRequired(t *testing.T) {
	fs, _, _ := testFlags()
	assert.DeepEqual(t, MissingRequired(fs), []string{"bar"})
	fs.Parse([]string{"--bar=baz"})
	assert.DeepEqual(t, MissingRequired(fs), []string{})
}

func TestAllRequired(t *testing.T) {
	fs, _, _ := testFlags()
	assert.DeepEqual(t, AllRequired(fs), []string{"bar"})
	fs.Parse([]string{"--bar=baz"})
	assert.DeepEqual(t, AllRequired(fs), []string{"bar"})
}

func TestEnumerateNil(t *testing.T) {
	got := Enumerate(nil)
	assert.Equal(t, len(got), 0)
}

func TestEnumerateEmpty(t *testing.T) {
	got := Enumerate(&flag.FlagSet{})
	assert.Equal(t, len(got), 0)
}

func TestEnumerate(t *testing.T) {
	fs, _, _ := testFlags()
	got := Enumerate(fs)
	assert.Equal(t, len(got), 2)
	assert.True(t,
		(got[0].Name == "foo-baz" && got[1].Name == "bar") ||
			(got[0].Name == "bar" && got[1].Name == "foo-baz"),
	)
}
