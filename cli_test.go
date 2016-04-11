package cli

import (
	"flag"
	"testing"

	"github.com/turbinelabs/tbn/cli/command"
	"github.com/turbinelabs/tbn/test/assert"
)

type testRunner struct {
	t      *testing.T
	name   string
	args   []string
	cmdErr command.CmdErr
	called int
}

type badInputTestCase struct {
	args              []string
	msg               string
	fillFlagsCalled   int
	usageGlobalCalled int
	usageCmdCalled    int
}

func (r *testRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	assert.Equal(r.t, cmd.Name, r.name)
	assert.DeepEqual(r.t, args, r.args)
	r.called++
	return r.cmdErr
}

type cliTestMock struct {
	t    *testing.T
	cmds []*command.Cmd
	args []string

	wantUsageFlagSet    *flag.FlagSet
	wantUsageGlobalCmds []*command.Cmd
	wantUsageCmd        *command.Cmd
	wantFillFlagSets    []*flag.FlagSet
	wantMsg             string
	wantCode            int

	usageGlobalCalled  int
	usageCmdCalled     int
	versionPrintCalled int
	fillFlagsCalled    int
	stderrCalled       int
	exitCalled         int

	barFlag bool
}

func (m *cliTestMock) Global(cmds []*command.Cmd, fs *flag.FlagSet) {
	assert.DeepEqual(m.t, cmds, m.wantUsageGlobalCmds)
	assert.DeepEqual(m.t, fs, m.wantUsageFlagSet)
	m.usageGlobalCalled++
}

func (m *cliTestMock) Command(cmd *command.Cmd) {
	assert.Equal(m.t, cmd, m.wantUsageCmd)
	m.usageCmdCalled++
}

func (m *cliTestMock) Print() {
	m.versionPrintCalled++
}

func (m *cliTestMock) fill(fs *flag.FlagSet) {
	if len(m.wantFillFlagSets) == 0 {
		m.t.Errorf("unexpected flagset: %v", fs)
		return
	}

	want := m.wantFillFlagSets[0]
	if len(m.wantFillFlagSets) > 0 {
		m.wantFillFlagSets = m.wantFillFlagSets[1:]
	}

	assert.DeepEqual(m.t, fs, want)

	m.fillFlagsCalled++
}

func (m *cliTestMock) osArgs() []string {
	return append([]string{"appname"}, m.args...)
}

func (m *cliTestMock) stderr(msg string) {
	assert.Equal(m.t, msg, m.wantMsg)
	m.stderrCalled++
}

func (m *cliTestMock) exit(code int) {
	assert.Equal(m.t, code, m.wantCode)
	m.exitCalled++
}

func newCliAndMock(t *testing.T, hasSubs bool) (*CLI, *cliTestMock) {
	cmds := []*command.Cmd{{
		Name:   "foo",
		Runner: &testRunner{t: t, name: "foo", args: []string{}, cmdErr: command.NoError()},
	}}

	if hasSubs {
		cmds = append(cmds, &command.Cmd{
			Name:   "bar",
			Runner: &testRunner{t: t, name: "bar", args: []string{}, cmdErr: command.Error("Gah!")},
		})
	}

	m := &cliTestMock{t: t, cmds: cmds}

	cli := &CLI{
		commands: m.cmds,
		usage:    m,
		version:  m,

		fillFlagsFromEnv: m.fill,
		osArgs:           m.osArgs,
		stderr:           m.stderr,
		exit:             m.exit,
	}

	m.wantUsageGlobalCmds = cli.commands
	m.wantUsageFlagSet = &cli.Flags
	m.wantFillFlagSets = []*flag.FlagSet{&cli.Flags}

	return cli, m
}

var (
	subBadInputTestCases = []badInputTestCase{
		{nil, "no command specified\n\n", 1, 1, 0},
		{[]string{}, "no command specified\n\n", 1, 1, 0},
		{[]string{"bogus"}, "unknown command: \"bogus\"\n\n", 1, 1, 0},
		{[]string{"bogus", "-version"}, "unknown command: \"bogus\"\n\n", 1, 1, 0},
		{[]string{"-bogus"}, "flag provided but not defined: -bogus\n\n", 0, 1, 0},
		{[]string{"foo", "-bogus"}, "foo: flag provided but not defined: -bogus\n\n", 1, 0, 1},
	}

	subGlobalHelpTestCases = [][]string{
		{"help"},
		{"-help"},
		{"-h"},
		{"help", "bogus"},
		{"-help", "bogus"},
		{"-h", "bogus"},
		{"bogus", "-help"},
		{"bogus", "-h"},
	}

	subCmdHelpTestCases = [][]string{
		{"help", "foo"},
		{"-help", "foo"},
		{"-h", "foo"},
		{"foo", "-help"},
		{"foo", "-h"},
	}

	subVersionTestCases = [][]string{
		{"version"},
		{"-version"},
		{"-v"},
		{"version", "foo"},
		{"-version", "foo"},
		{"-v", "foo"},
		{"version", "bogus"},
		{"-version", "bogus"},
		{"-v", "bogus"},
		{"foo", "-version"},
		{"foo", "-v"},
	}

	badInputTestCases = []badInputTestCase{
		{[]string{"-bogus"}, "foo: flag provided but not defined: -bogus\n\n", 0, 0, 1},
	}

	helpTestCases = [][]string{
		{"-help"},
		{"-h"},
	}

	versionTestCases = [][]string{
		{"-version"},
		{"-v"},
	}
)

func TestBadInput(t *testing.T) {
	cli, mock := newCliAndMock(t, false)
	mock.args = []string{"-bogus"}
	mock.wantCode = 2
	mock.wantMsg = "foo: flag provided but not defined: -bogus\n\n"
	mock.wantUsageCmd = mock.cmds[0]

	cli.Main()
	assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.versionPrintCalled, 0)
	assert.Equal(t, mock.fillFlagsCalled, 0)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 1)
	assert.Equal(t, mock.stderrCalled, 1)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestHelp(t *testing.T) {
	for _, tc := range helpTestCases {
		cli, mock := newCliAndMock(t, false)
		mock.args = tc

		mock.wantCode = 0
		mock.wantFillFlagSets = []*flag.FlagSet{&mock.cmds[0].Flags}
		mock.wantUsageCmd = mock.cmds[0]

		cli.Main()
		assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 0)
		assert.Equal(t, mock.versionPrintCalled, 0)
		assert.Equal(t, mock.fillFlagsCalled, 1)
		assert.Equal(t, mock.usageGlobalCalled, 0)
		assert.Equal(t, mock.usageCmdCalled, 1)
		assert.Equal(t, mock.stderrCalled, 0)
		assert.Equal(t, mock.exitCalled, 1)
	}
}

func TestVersion(t *testing.T) {
	for _, tc := range versionTestCases {
		cli, mock := newCliAndMock(t, false)
		mock.args = tc

		mock.wantCode = 0
		mock.wantFillFlagSets = []*flag.FlagSet{&mock.cmds[0].Flags}

		cli.Main()
		assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 0)
		assert.Equal(t, mock.versionPrintCalled, 1)
		assert.Equal(t, mock.fillFlagsCalled, 1)
		assert.Equal(t, mock.usageGlobalCalled, 0)
		assert.Equal(t, mock.usageCmdCalled, 0)
		assert.Equal(t, mock.stderrCalled, 0)
		assert.Equal(t, mock.exitCalled, 1)
	}
}

func TestNoError(t *testing.T) {
	cli, mock := newCliAndMock(t, false)
	mock.args = []string{"-bar", "baz"}
	mock.cmds[0].Flags.BoolVar(&mock.barFlag, "bar", false, "")
	mock.cmds[0].Runner.(*testRunner).args = []string{"baz"}

	mock.wantCode = 0
	mock.wantFillFlagSets = []*flag.FlagSet{&mock.cmds[0].Flags}

	cli.Main()
	assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 1)
	assert.Equal(t, mock.versionPrintCalled, 0)
	assert.Equal(t, mock.fillFlagsCalled, 1)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestError(t *testing.T) {
	cli, mock := newCliAndMock(t, false)
	mock.cmds[0].Runner.(*testRunner).cmdErr = mock.cmds[0].Error("das ist borken!")

	mock.wantCode = 1
	mock.wantMsg = "foo: das ist borken!\n\n"
	mock.wantFillFlagSets = []*flag.FlagSet{&mock.cmds[0].Flags}

	cli.Main()
	assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 1)
	assert.Equal(t, mock.versionPrintCalled, 0)
	assert.Equal(t, mock.fillFlagsCalled, 1)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 1)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestSubBadInput(t *testing.T) {
	for _, tc := range subBadInputTestCases {
		cli, mock := newCliAndMock(t, true)

		mock.args = tc.args
		mock.wantCode = 2
		mock.wantMsg = tc.msg
		if tc.fillFlagsCalled > 0 {
			mock.wantUsageCmd = mock.cmds[0]
		}

		cli.Main()
		assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 0)
		assert.Equal(t, mock.versionPrintCalled, 0)
		assert.Equal(t, mock.fillFlagsCalled, tc.fillFlagsCalled)
		assert.Equal(t, mock.usageGlobalCalled, tc.usageGlobalCalled)
		assert.Equal(t, mock.usageCmdCalled, tc.usageCmdCalled)
		assert.Equal(t, mock.stderrCalled, 1)
		assert.Equal(t, mock.exitCalled, 1)
	}
}

func TestSubGlobalHelp(t *testing.T) {
	for _, tc := range subGlobalHelpTestCases {
		cli, mock := newCliAndMock(t, true)
		mock.args = tc

		mock.wantCode = 0

		cli.Main()
		assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 0)
		assert.Equal(t, mock.versionPrintCalled, 0)
		assert.Equal(t, mock.fillFlagsCalled, 1)
		assert.Equal(t, mock.usageGlobalCalled, 1)
		assert.Equal(t, mock.usageCmdCalled, 0)
		assert.Equal(t, mock.stderrCalled, 0)
		assert.Equal(t, mock.exitCalled, 1)
	}
}

func TestSubCmdHelp(t *testing.T) {
	for _, tc := range subCmdHelpTestCases {
		cli, mock := newCliAndMock(t, true)
		mock.args = tc

		mock.wantCode = 0
		mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.cmds[0].Flags}
		mock.wantUsageCmd = mock.cmds[0]

		cli.Main()
		assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 0)
		assert.Equal(t, mock.versionPrintCalled, 0)
		assert.Equal(t, mock.fillFlagsCalled, 2)
		assert.Equal(t, mock.usageGlobalCalled, 0)
		assert.Equal(t, mock.usageCmdCalled, 1)
		assert.Equal(t, mock.stderrCalled, 0)
		assert.Equal(t, mock.exitCalled, 1)
	}
}

func TestSubVersion(t *testing.T) {
	for _, tc := range subVersionTestCases {
		cli, mock := newCliAndMock(t, true)
		mock.args = tc

		mock.wantCode = 0
		if tc[0] == "foo" {
			mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.cmds[0].Flags}
		}

		cli.Main()
		assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 0)
		assert.Equal(t, mock.versionPrintCalled, 1)
		if tc[0] == "foo" {
			assert.Equal(t, mock.fillFlagsCalled, 2)
		} else {
			assert.Equal(t, mock.fillFlagsCalled, 1)
		}
		assert.Equal(t, mock.usageGlobalCalled, 0)
		assert.Equal(t, mock.usageCmdCalled, 0)
		assert.Equal(t, mock.stderrCalled, 0)
		assert.Equal(t, mock.exitCalled, 1)
	}
}

func TestSubFoo(t *testing.T) {
	cli, mock := newCliAndMock(t, true)
	mock.args = []string{"foo", "-bar", "baz"}
	mock.cmds[0].Flags.BoolVar(&mock.barFlag, "bar", false, "")
	mock.cmds[0].Runner.(*testRunner).args = []string{"baz"}

	mock.wantCode = 0
	mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.cmds[0].Flags}

	cli.Main()
	assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 1)
	assert.Equal(t, mock.versionPrintCalled, 0)
	assert.Equal(t, mock.fillFlagsCalled, 2)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestSubFooError(t *testing.T) {
	cli, mock := newCliAndMock(t, true)
	mock.args = []string{"foo"}
	mock.cmds[0].Runner.(*testRunner).cmdErr = mock.cmds[0].Error("das ist borken!")

	mock.wantCode = 1
	mock.wantMsg = "foo: das ist borken!\n\n"
	mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.cmds[0].Flags}

	cli.Main()
	assert.Equal(t, mock.cmds[0].Runner.(*testRunner).called, 1)
	assert.Equal(t, mock.versionPrintCalled, 0)
	assert.Equal(t, mock.fillFlagsCalled, 2)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 1)
	assert.Equal(t, mock.exitCalled, 1)
}
