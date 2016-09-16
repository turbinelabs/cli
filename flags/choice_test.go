package flags

import (
	"testing"

	"github.com/turbinelabs/test/assert"
)

func TestNewChoice(t *testing.T) {
	c := NewChoice([]string{"a", "b", "c"})
	assert.Nil(t, c.Choice)
	assert.DeepEqual(t, c.AllowedValues, []string{"a", "b", "c"})
}

func TestChoiceString(t *testing.T) {
	value := "v1"
	c := &Choice{Choice: &value}
	assert.Equal(t, c.String(), value)
}

func TestChoiceGet(t *testing.T) {
	value := "v1"
	c := &Choice{Choice: &value}
	assert.DeepEqual(t, c.Get(), &value)
}

func TestChoiceSet(t *testing.T) {
	v := "b"
	c := &Choice{AllowedValues: []string{"a", "b", "c"}}
	err := c.Set("b")
	assert.DeepEqual(t, c.Choice, &v)
	assert.Nil(t, err)

	c.Choice = nil
	err = c.Set("nope")
	assert.Nil(t, c.Choice)
	assert.ErrorContains(t, err, "invalid flag value: nope, must be one of a, b, c")
}
