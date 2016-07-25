package flags

import (
	"flag"
	"testing"

	"github.com/golang/mock/gomock"

	tbnos "github.com/turbinelabs/os"
	"github.com/turbinelabs/test/assert"
)

func TestFromEnvPrefix(t *testing.T) {
	want := "A_B_CD_A_B_CD_A_B_CD"
	values := []string{
		"A__B_CD",
		"A-.B*CD",
		"A.b-##cd",
		"a()&b-cd",
		"aö@bë^cD",
		"a\t    b\ncD",
	}
	for _, v := range values {
		assert.Equal(t, EnvKey(v, v, v), want)
		assert.Equal(t, NewFromEnv(nil, v, v, v).Prefix(), want+"_")
	}
}

func TestFromEnvAllFlagsNil(t *testing.T) {
	got := NewFromEnv(nil).AllFlags()
	assert.Equal(t, len(got), 0)
}

func TestFromEnvAllFlagsEmpty(t *testing.T) {
	got := NewFromEnv(&flag.FlagSet{}).AllFlags()
	assert.Equal(t, len(got), 0)
}

func TestFromEnvAllFlags(t *testing.T) {
	fs, _, _ := testFlags()
	got := NewFromEnv(fs).AllFlags()
	assert.Equal(t, len(got), 2)
	assert.True(t,
		(got[0].Name == "foo-baz" && got[1].Name == "bar") ||
			(got[0].Name == "bar" && got[1].Name == "foo-baz"),
	)
}

func TestFillFromEnvAllUnset(t *testing.T) {
	ctrl := gomock.NewController(assert.Tracing(t))
	defer ctrl.Finish()

	mockOS := tbnos.NewMockOS(ctrl)
	mockOS.EXPECT().Getenv("FOO_BAR_FOO_BAZ").Return("blegga")
	mockOS.EXPECT().Getenv("FOO_BAR_BAR").Return("")

	fs, fooFlag, barFlag := testFlags()
	fs.Parse([]string{})

	fe := NewFromEnv(fs, "foo", "bar").(fromEnv)
	fe.os = mockOS

	fe.Fill()

	assert.Equal(t, *fooFlag, "blegga")
	assert.Equal(t, *barFlag, "")
	assert.DeepEqual(t, fe.Filled(), map[string]string{"FOO_BAR_FOO_BAZ": "blegga"})
}

func TestFillFromEnvOneSet(t *testing.T) {
	ctrl := gomock.NewController(assert.Tracing(t))
	defer ctrl.Finish()

	mockOS := tbnos.NewMockOS(ctrl)
	mockOS.EXPECT().Getenv("FOO_BAR_BAR").Return("")

	fs, fooFlag, barFlag := testFlags()
	fs.Parse([]string{"--foo-baz=blargo"})

	fe := NewFromEnv(fs, "foo", "bar").(fromEnv)
	fe.os = mockOS

	fe.Fill()

	assert.Equal(t, *fooFlag, "blargo")
	assert.Equal(t, *barFlag, "")
	assert.DeepEqual(t, fe.Filled(), map[string]string{})
}
