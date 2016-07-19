package flags

import (
	"flag"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/turbinelabs/test/assert"
)

const (
	flagUsage    = "a flag for {{NAME}}"
	flagUsageFmt = "a flag for %s"
)

type prefixedFlagTestCase struct {
	name     string
	flagType reflect.Type

	addFlag func(f *PrefixedFlagSet) interface{}
}

type stringValue struct {
	value string
}

func (sv *stringValue) Set(s string) error {
	sv.value = s
	return nil
}

func (sv *stringValue) String() string {
	return sv.value
}

var _ flag.Value = &stringValue{}

func (tc *prefixedFlagTestCase) run(t testing.TB) {
	fs := flag.NewFlagSet("generated code: "+tc.name, flag.PanicOnError)

	pfs := NewPrefixedFlagSet(fs, "theprefix", "the-app-name")

	target := tc.addFlag(pfs)

	var value string
	var expectedValue interface{}

	switch tc.flagType.Kind() {
	case reflect.Bool:
		value = "true"
		ev := true
		expectedValue = &ev
	case reflect.Int:
		value = "123"
		ev := 123
		expectedValue = &ev
	case reflect.Int64:
		if tc.flagType == reflect.TypeOf(time.Duration(0)) {
			value = "10s"
			ev := time.Duration(10 * time.Second)
			expectedValue = &ev
		} else {
			value = "123"
			ev := int64(123)
			expectedValue = &ev
		}
	case reflect.Uint:
		value = "123"
		ev := uint(123)
		expectedValue = &ev
	case reflect.Uint64:
		value = "123"
		ev := uint64(123)
		expectedValue = &ev
	case reflect.Float64:
		value = "12.3"
		ev := float64(12.3)
		expectedValue = &ev
	case reflect.String:
		value = "something"
		expectedValue = &value
	default:
		t.Fatalf("unhandled type %s", tc.flagType.String())
		return
	}

	flagName := "theprefix." + tc.name

	pfs.Parse([]string{
		"-" + flagName + "=" + value,
	})

	if valueTarget, ok := target.(flag.Value); ok {
		v := valueTarget.String()
		assert.DeepEqual(t, &v, expectedValue)
	} else {
		assert.DeepEqual(t, target, expectedValue)
	}

	f := pfs.Lookup(flagName)
	assert.NonNil(t, f)
	assert.Equal(t, f.Name, flagName)
	assert.Equal(t, f.Usage, fmt.Sprintf(flagUsageFmt, "the-app-name"))
}

func TestGeneratedCode(t *testing.T) {
	for _, tc := range generatedTestCases {
		assert.Group(fmt.Sprintf("for flag type %s", tc.name), t, func(g *assert.G) {
			tc.run(g)
		})
	}
}

func TestVar(t *testing.T) {
	testCase := prefixedFlagTestCase{
		name:     "var",
		flagType: reflect.TypeOf(string(0)),
		addFlag: func(f *PrefixedFlagSet) interface{} {
			var target stringValue
			f.Var(&target, "var", "a flag for {{NAME}}")
			return &target
		},
	}
	testCase.run(t)
}
