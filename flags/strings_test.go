package flags

import (
	"testing"

	"github.com/turbinelabs/test/assert"
)

func TestNewStrings(t *testing.T) {
	s := NewStrings()
	assert.Nil(t, s.Strings)
	assert.Nil(t, s.AllowedValues)
	assert.Equal(t, s.Delimiter, ",")
}

func TestNewStringsWithConstraint(t *testing.T) {
	s := NewStringsWithConstraint([]string{"a", "b", "c"})
	assert.Nil(t, s.Strings)
	assert.DeepEqual(t, s.AllowedValues, []string{"a", "b", "c"})
	assert.Equal(t, s.Delimiter, ",")
}

func TestStringsString(t *testing.T) {
	s := &Strings{Strings: []string{"v1", "v2"}, Delimiter: "-"}
	assert.Equal(t, s.String(), "v1-v2")
}

func TestStringGet(t *testing.T) {
	s := &Strings{Strings: []string{"v1", "v2"}, Delimiter: ","}
	assert.DeepEqual(t, s.Get(), s.Strings)
}

func TestStringSet(t *testing.T) {
	s := &Strings{Delimiter: ","}
	err := s.Set("a,b,c")
	assert.DeepEqual(t, s.Strings, []string{"a", "b", "c"})
	assert.Nil(t, err)

	err = s.Set("a")
	assert.DeepEqual(t, s.Strings, []string{"a"})
	assert.Nil(t, err)

	err = s.Set(",,,,,,")
	assert.DeepEqual(t, s.Strings, []string{})
	assert.Nil(t, err)

	err = s.Set(",,a,,c,,")
	assert.DeepEqual(t, s.Strings, []string{"a", "c"})
	assert.Nil(t, err)

	err = s.Set(" , , c , , a , , ")
	assert.DeepEqual(t, s.Strings, []string{"c", "a"})
	assert.Nil(t, err)

	err = s.Set("")
	assert.DeepEqual(t, s.Strings, []string{})
	assert.Nil(t, err)

	s = &Strings{Delimiter: "-"}
	err = s.Set("a-b-c")
	assert.DeepEqual(t, s.Strings, []string{"a", "b", "c"})
	assert.Nil(t, err)

	err = s.Set("a,b,c")
	assert.DeepEqual(t, s.Strings, []string{"a,b,c"})
	assert.Nil(t, err)
}

func TestStringSetWithConstraint(t *testing.T) {
	s := &Strings{AllowedValues: []string{"a", "b", "c"}, Delimiter: ","}
	err := s.Set("a,b,c")
	assert.DeepEqual(t, s.Strings, []string{"a", "b", "c"})
	assert.Nil(t, err)

	s.Strings = []string{}

	err = s.Set("a,b,c,d")
	assert.DeepEqual(t, s.Strings, []string{})
	assert.ErrorContains(t, err, "invalid flag value(s): d")

	err = s.Set("x,a,y,b,z,c")
	assert.DeepEqual(t, s.Strings, []string{})
	assert.ErrorContains(t, err, "invalid flag value(s): x, y, z")
}
