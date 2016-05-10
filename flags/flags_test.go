package flags

import (
	"flag"
	"testing"

	"github.com/turbinelabs/test/assert"
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
	var fs flag.FlagSet
	fooFlag := fs.String("foo-baz", "", "do the foo")
	barFlag := fs.String("bar", "", "harty har to the bar")
	return &fs, fooFlag, barFlag
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

func TestFillFromEnvAllUnset(t *testing.T) {
	mock := flagsTestMock{}
	fs, fooFlag, barFlag := testFlags()
	fs.Parse([]string{})
	fillFromEnv("foo-bar", fs, mock.getenv)
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
	fs, fooFlag, barFlag := testFlags()
	fs.Parse([]string{"--foo-baz=blargo"})
	fillFromEnv("foo-bar", fs, mock.getenv)
	assert.Equal(t, len(mock.keys), 1)
	assert.True(t, mock.keys[0] == "FOO_BAR_BAR")
	assert.Equal(t, *fooFlag, "blargo")
	assert.Equal(t, *barFlag, "")
}

func TestEnvKey(t *testing.T) {
	want := "A_B_CD_A_B_CD"
	values := []string{
		"A_B_CD",
		"A-B*CD",
		"A.b-cd",
		"a&b-cd",
		"aöbëcD",
		"a\tb\ncD",
	}
	for _, v := range values {
		assert.Equal(t, EnvKey(v, v), want)
	}
}
