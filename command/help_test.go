package command

import (
	"testing"

	"github.com/turbinelabs/tbn/testhelpers/assert"
)

var helpTestCmd = &Cmd{Name: "foo"}

type helpTestMock struct {
	globalUsages int
	cmdUsages    []string
	commands     []string
}

func (u *helpTestMock) command(name string) *Cmd {
	u.commands = append(u.commands, name)
	if name == "foo" {
		return helpTestCmd
	}
	return nil
}

func (u *helpTestMock) globalUsage() {
	u.globalUsages += 1
}

func (u *helpTestMock) cmdUsage(cmd *Cmd) {
	u.cmdUsages = append(u.cmdUsages, cmd.Name)
}

func (u *helpTestMock) newHelpRunner() *helpRunner {
	return &helpRunner{
		u.command,
		u.globalUsage,
		u.cmdUsage,
	}
}

func TestRunNilArgs(t *testing.T) {
	mock := helpTestMock{}
	runner := mock.newHelpRunner()
	err := runner.Run(helpTestCmd, nil)
	assert.Equal(t, err, NoError())
	assert.Equal(t, mock.globalUsages, 1)
	assert.Equal(t, len(mock.cmdUsages), 0)
	assert.Equal(t, len(mock.commands), 0)
}

func TestRunEmptyArgs(t *testing.T) {
	mock := helpTestMock{}
	runner := mock.newHelpRunner()
	err := runner.Run(helpTestCmd, []string{})
	assert.Equal(t, err, NoError())
	assert.Equal(t, mock.globalUsages, 1)
	assert.Equal(t, len(mock.cmdUsages), 0)
	assert.Equal(t, len(mock.commands), 0)
}

func TestRunOneGoodArg(t *testing.T) {
	mock := helpTestMock{}
	runner := mock.newHelpRunner()
	err := runner.Run(helpTestCmd, []string{"foo"})
	assert.Equal(t, err, NoError())
	assert.Equal(t, mock.globalUsages, 0)
	assert.Equal(t, len(mock.cmdUsages), 1)
	assert.Equal(t, len(mock.commands), 1)
	assert.Equal(t, mock.cmdUsages[0], "foo")
	assert.Equal(t, mock.commands[0], "foo")
}

func TestRunOneBadArg(t *testing.T) {
	mock := helpTestMock{}
	runner := mock.newHelpRunner()
	err := runner.Run(helpTestCmd, []string{"bar"})
	assert.Equal(t, err, helpTestCmd.BadInput(`unknown command: "bar"`))
	assert.Equal(t, mock.globalUsages, 0)
	assert.Equal(t, len(mock.cmdUsages), 0)
	assert.Equal(t, len(mock.commands), 1)
	assert.Equal(t, mock.commands[0], "bar")
}

func TestRunIgnoreExtraArgs(t *testing.T) {
	mock := helpTestMock{}
	runner := mock.newHelpRunner()
	err := runner.Run(helpTestCmd, []string{"foo", "bar"})
	assert.Equal(t, err, NoError())
	assert.Equal(t, mock.globalUsages, 0)
	assert.Equal(t, len(mock.cmdUsages), 1)
	assert.Equal(t, len(mock.commands), 1)
	assert.Equal(t, mock.cmdUsages[0], "foo")
	assert.Equal(t, mock.commands[0], "foo")
}
