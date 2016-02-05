package command

import (
	"flag"
	"testing"

	"github.com/turbinelabs/tbn/testhelpers/assert"
)

type testRunner struct {
	name string
	args []string
	t    *testing.T
}

func (r testRunner) Run(cmd *Cmd, args []string) CmdErr {
	assert.Equal(r.t, cmd.Name, r.name)
	assert.DeepEqual(r.t, args, r.args)
	return cmd.Error("baz")
}

func TestCmdRun(t *testing.T) {
	var flags flag.FlagSet
	flags.Parse([]string{"foo"})

	cmd := Cmd{Name: "bar", Flags: flags, Runner: testRunner{"bar", []string{"foo"}, t}}
	err := cmd.Run()
	assert.Equal(t, err, CmdErr{&cmd, CmdErrCodeError, "bar: baz"})
}

func TestCmdBadInputf(t *testing.T) {
	cmd := Cmd{Name: "bar"}
	want := CmdErr{&cmd, CmdErrCodeBadInput, "bar: 1-2-3"}
	got := cmd.BadInputf("%d-%d-%d", 1, 2, 3)
	assert.Equal(t, got, want)
}

func TestCmdErrorf(t *testing.T) {
	cmd := Cmd{Name: "bar"}
	want := CmdErr{&cmd, CmdErrCodeError, "bar: 1-2-3"}
	got := cmd.Errorf("%d-%d-%d", 1, 2, 3)
	assert.Equal(t, got, want)
}

func TestCmdBadInput(t *testing.T) {
	cmd := Cmd{Name: "bar"}
	want := CmdErr{&cmd, CmdErrCodeBadInput, "bar: baz:1 2 3"}
	got := cmd.BadInput("baz:", 1, 2, 3)
	assert.Equal(t, got, want)
}

func TestCmdError(t *testing.T) {
	cmd := Cmd{Name: "bar"}
	want := CmdErr{&cmd, CmdErrCodeError, "bar: baz:1 2 3"}
	got := cmd.Error("baz:", 1, 2, 3)
	assert.Equal(t, got, want)
}

func TestIsError(t *testing.T) {
	assert.True(t, BadInput("").IsError())
	assert.True(t, Error("").IsError())
	assert.False(t, NoError().IsError())
}

func TestBadInputf(t *testing.T) {
	want := CmdErr{nil, CmdErrCodeBadInput, "1-2-3"}
	got := BadInputf("%d-%d-%d", 1, 2, 3)
	assert.Equal(t, got, want)
}

func TestErrorf(t *testing.T) {
	want := CmdErr{nil, CmdErrCodeError, "1-2-3"}
	got := Errorf("%d-%d-%d", 1, 2, 3)
	assert.Equal(t, got, want)
}

func TestBadInput(t *testing.T) {
	want := CmdErr{nil, CmdErrCodeBadInput, "baz:1 2 3"}
	got := BadInput("baz:", 1, 2, 3)
	assert.Equal(t, got, want)
}

func TestError(t *testing.T) {
	want := CmdErr{nil, CmdErrCodeError, "baz:1 2 3"}
	got := Error("baz:", 1, 2, 3)
	assert.Equal(t, got, want)
}

func TestNoError(t *testing.T) {
	want := CmdErr{nil, CmdErrCodeNoError, ""}
	got := NoError()
	assert.Equal(t, got, want)
}
