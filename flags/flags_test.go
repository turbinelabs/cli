package flags

import (
	"flag"
	"testing"

	"github.com/turbinelabs/tbn/testhelpers/assert"
)

type flagsTestMock struct {
	keys []string
}

func (f *flagsTestMock) getenv(key string) string {
	f.keys = append(f.keys, key)
	if key == "FOO_BAR_FOO_BAZ" {
		return "blegga"
	}
	return ""
}

func testFlags() (*flag.FlagSet, *string, *string) {
	var flags flag.FlagSet
	fooFlag := flags.String("foo-baz", "", "do the foo")
	barFlag := flags.String("bar", "", "harty har to the bar")
	return &flags, fooFlag, barFlag
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
	flags, _, _ := testFlags()
	got := Enumerate(flags)
	assert.Equal(t, len(got), 2)
	assert.True(t,
		(got[0].Name == "foo-baz" && got[1].Name == "bar") ||
			(got[0].Name == "bar" && got[1].Name == "foo-baz"),
	)
}

func TestFillFromEnvAllUnset(t *testing.T) {
	mock := flagsTestMock{}
	flags, fooFlag, barFlag := testFlags()
	flags.Parse([]string{})
	fillFromEnv("foo-bar", flags, mock.getenv)
	assert.Equal(t, len(mock.keys), 2)
	assert.True(t,
		(mock.keys[0] == "FOO_BAR_FOO_BAZ" && mock.keys[1] == "FOO_BAR_BAR") ||
			(mock.keys[0] == "FOO_BAR_BAR" && mock.keys[1] == "FOO_BAR_FOO_BAZ"),
	)
	assert.Equal(t, *fooFlag, "blegga")
	assert.Equal(t, *barFlag, "")
}

func TestFillFromEnvOneSet(t *testing.T) {
	mock := flagsTestMock{}
	flags, fooFlag, barFlag := testFlags()
	flags.Parse([]string{"--foo-baz=blargo"})
	fillFromEnv("foo-bar", flags, mock.getenv)
	assert.Equal(t, len(mock.keys), 1)
	assert.True(t, mock.keys[0] == "FOO_BAR_BAR")
	assert.Equal(t, *fooFlag, "blargo")
	assert.Equal(t, *barFlag, "")
}
