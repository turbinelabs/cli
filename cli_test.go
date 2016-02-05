package cli

import (
	"flag"
	"testing"

	"github.com/turbinelabs/tbn/cli/app"
	"github.com/turbinelabs/tbn/cli/command"
	"github.com/turbinelabs/tbn/testhelpers/assert"
)

type testRunner struct {
	t      *testing.T
	name   string
	args   []string
	cmdErr command.CmdErr
	called int
}

func (r *testRunner) Run(cmd *command.Cmd, args []string) command.CmdErr {
	assert.Equal(r.t, cmd.Name, r.name)
	assert.DeepEqual(r.t, args, r.args)
	r.called += 1
	return r.cmdErr
}

type cliTestMock struct {
	t              *testing.T
	commandFoo     *command.Cmd
	commandHelp    *command.Cmd
	commandVersion *command.Cmd
	args           []string

	wantUsageFlagSet    *flag.FlagSet
	wantUsageGlobalCmds []*command.Cmd
	wantUsageCmd        *command.Cmd
	wantFillFlagSets    []*flag.FlagSet
	wantMsg             string
	wantCode            int

	usageGlobalCalled int
	usageCmdCalled    int
	fillFlagsCalled   int
	stderrCalled      int
	exitCalled        int

	barFlag bool
}

func (m *cliTestMock) Global(cmds []*command.Cmd, fs *flag.FlagSet) {
	assert.DeepEqual(m.t, cmds, m.wantUsageGlobalCmds)
	assert.DeepEqual(m.t, fs, m.wantUsageFlagSet)
	m.usageGlobalCalled += 1
}

func (m *cliTestMock) Command(cmd *command.Cmd) {
	assert.Equal(m.t, cmd, m.wantUsageCmd)
	m.usageCmdCalled += 1
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

	m.fillFlagsCalled += 1
}

func (m *cliTestMock) osArgs() []string {
	return append([]string{"appname"}, m.args...)
}

func (m *cliTestMock) stderr(msg string) {
	assert.Equal(m.t, msg, m.wantMsg)
	m.stderrCalled += 1
}

func (m *cliTestMock) exit(code int) {
	assert.Equal(m.t, code, m.wantCode)
	m.exitCalled += 1
}

func newCliAndMock(t *testing.T) (*CLI, *cliTestMock) {
	m := &cliTestMock{
		t: t,
		commandFoo: &command.Cmd{
			Name:   "foo",
			Runner: &testRunner{t: t, name: "foo", args: []string{}, cmdErr: command.NoError()},
		},
		commandHelp: &command.Cmd{
			Name:   "help",
			Runner: &testRunner{t: t, name: "help", args: []string{}, cmdErr: command.NoError()},
		},
		commandVersion: &command.Cmd{
			Name:   "version",
			Runner: &testRunner{t: t, name: "version", args: []string{}, cmdErr: command.NoError()},
		},
	}

	cli := new(CLI)
	cli.commands = []*command.Cmd{m.commandFoo, m.commandHelp, m.commandVersion}

	cli.fillFlagsFromEnv = m.fill
	cli.osArgs = m.osArgs
	cli.usage = m
	cli.stderr = m.stderr
	cli.exit = m.exit

	addVersionFlag(&cli.Flags, &cli.versionFlag)
	addHelpFlag(&cli.Flags, &cli.helpFlag)

	m.wantUsageGlobalCmds = cli.commands
	m.wantUsageFlagSet = &cli.Flags
	m.wantFillFlagSets = []*flag.FlagSet{&cli.Flags}

	return cli, m
}

func TestNewCLIAddsHelpAndVersion(t *testing.T) {
	cli := NewCLI(app.App{}, &command.Cmd{Name: "foo"})
	assert.Equal(t, len(cli.commands), 3)
	assert.Equal(t, cli.commands[0].Name, "foo")
	assert.Equal(t, cli.commands[1].Name, "help")
	assert.Equal(t, cli.commands[2].Name, "version")
}

func TestCLIBadGlobalFlags(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"-bogus"}

	mock.wantCode = 2
	mock.wantMsg = "flag provided but not defined: -bogus\n\n"

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 0)
	assert.Equal(t, mock.usageGlobalCalled, 1)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 1)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIUnknownCommand(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"bogus"}

	mock.wantCode = 2
	mock.wantMsg = "unknown command: \"bogus\"\n\n"

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 1)
	assert.Equal(t, mock.usageGlobalCalled, 1)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 1)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLINoArgs(t *testing.T) {
	cli, mock := newCliAndMock(t)

	mock.wantCode = 2
	mock.wantMsg = "no command specified\n\n"
	mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.commandHelp.Flags}

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 1)
	assert.Equal(t, mock.usageGlobalCalled, 1)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 1)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIHelp(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"help", "foo"}
	mock.commandHelp.Runner.(*testRunner).args = []string{"foo"}

	mock.wantCode = 0
	mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.commandHelp.Flags}

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 1)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 2)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIHelpFlagNoArg(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"-help"}
	mock.commandHelp.Runner.(*testRunner).args = []string{"foo"}

	mock.wantCode = 0

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 1)
	assert.Equal(t, mock.usageGlobalCalled, 1)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIHelpFlagBeforeKnownCmd(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"-help", "foo"}

	mock.wantCode = 0
	mock.wantUsageCmd = mock.commandFoo

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 1)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 1)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIHelpFlagBeforeUnknownCmd(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"-help", "bar"}

	mock.wantCode = 0

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 1)
	assert.Equal(t, mock.usageGlobalCalled, 1)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIHelpFlagAfterKnownCmd(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"foo", "-help"}

	mock.wantCode = 0
	mock.wantUsageCmd = mock.commandFoo
	mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.commandFoo.Flags}

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 2)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 1)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIHelpFlagAfterUnknownCmd(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"bar", "-help"}

	mock.wantCode = 0

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 1)
	assert.Equal(t, mock.usageGlobalCalled, 1)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIVersion(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"version"}

	mock.wantCode = 0
	mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.commandVersion.Flags}

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 1)
	assert.Equal(t, mock.fillFlagsCalled, 2)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIVersionFlag(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"-version", "foo"}

	mock.wantCode = 0
	mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.commandVersion.Flags}

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 1)
	assert.Equal(t, mock.fillFlagsCalled, 2)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIFoo(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"foo", "-bar", "baz"}
	mock.commandFoo.Flags.BoolVar(&mock.barFlag, "bar", false, "")
	mock.commandFoo.Runner.(*testRunner).args = []string{"baz"}

	mock.wantCode = 0
	mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.commandFoo.Flags}

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 1)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 2)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 0)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIFooBadInput(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"foo", "-bogus"}

	mock.wantCode = 2
	mock.wantMsg = "foo: flag provided but not defined: -bogus\n\n"
	mock.wantUsageCmd = mock.commandFoo

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 1)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 1)
	assert.Equal(t, mock.stderrCalled, 1)
	assert.Equal(t, mock.exitCalled, 1)
}

func TestCLIFooError(t *testing.T) {
	cli, mock := newCliAndMock(t)
	mock.args = []string{"foo"}
	mock.commandFoo.Runner.(*testRunner).cmdErr = mock.commandFoo.Error("das ist borken!")

	mock.wantCode = 1
	mock.wantMsg = "foo: das ist borken!\n\n"
	mock.wantFillFlagSets = []*flag.FlagSet{&cli.Flags, &mock.commandFoo.Flags}

	cli.Main()
	assert.Equal(t, mock.commandFoo.Runner.(*testRunner).called, 1)
	assert.Equal(t, mock.commandHelp.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.commandVersion.Runner.(*testRunner).called, 0)
	assert.Equal(t, mock.fillFlagsCalled, 2)
	assert.Equal(t, mock.usageGlobalCalled, 0)
	assert.Equal(t, mock.usageCmdCalled, 0)
	assert.Equal(t, mock.stderrCalled, 1)
	assert.Equal(t, mock.exitCalled, 1)
}
