package cli

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/turbinelabs/cli/app"
	"github.com/turbinelabs/cli/command"
	"github.com/turbinelabs/cli/flags"
	tbnos "github.com/turbinelabs/os"
	"github.com/turbinelabs/test/assert"
)

func TestValidate(t *testing.T) {
	barCmd := &command.Cmd{Name: "bar"}
	barBazCmd := &command.Cmd{Name: "bar-baz"}
	fooCli := mkNew(app.App{Name: "foo"}, barCmd, barBazCmd)

	assert.Nil(t, fooCli.Validate())

	fooCli.Flags().Bool("bar-baz-blegga", false, "")
	barCmd.Flags.Bool("baz.blegga", false, "")
	barBazCmd.Flags.Bool("blegga", false, "")

	wantErr := errors.New(`possible environment key collisions:
  FOO_BAR_BAZ_BLEGGA: "foo -bar-baz-blegga", "foo bar -baz.blegga", "foo bar-baz -blegga"
`)

	assert.DeepEqual(t, fooCli.Validate(), wantErr)
}

type cmdType string

const (
	singleCmd    cmdType = "single"
	multipleCmds         = "multiple"
)

type cliTestMocks struct {
	t                  testing.TB
	fooRunner          *command.MockRunner
	barRunner          *command.MockRunner
	usage              *app.MockUsage
	version            *app.MockVersion
	flagsFromEnv       *flags.MockFromEnv
	cmdFooFlagsFromEnv *flags.MockFromEnv
	cmdBarFlagsFromEnv *flags.MockFromEnv
	os                 *tbnos.MockOS
	stderr             *bytes.Buffer
	finish             func()
}

func newCLIAndMocks(t testing.TB, cType cmdType) (*cli, *cliTestMocks) {
	ctrl := gomock.NewController(assert.Tracing(t))

	fooRunner := command.NewMockRunner(ctrl)
	barRunner := command.NewMockRunner(ctrl)

	cmds := []*command.Cmd{{Name: "foo", Runner: fooRunner}}

	if cType == multipleCmds {
		cmds = append(cmds, &command.Cmd{Name: "bar", Runner: barRunner})
	}

	mocks := &cliTestMocks{
		t:                  t,
		fooRunner:          fooRunner,
		barRunner:          barRunner,
		usage:              app.NewMockUsage(ctrl),
		version:            app.NewMockVersion(ctrl),
		flagsFromEnv:       flags.NewMockFromEnv(ctrl),
		cmdFooFlagsFromEnv: flags.NewMockFromEnv(ctrl),
		cmdBarFlagsFromEnv: flags.NewMockFromEnv(ctrl),
		os:                 tbnos.NewMockOS(ctrl),
		stderr:             &bytes.Buffer{},
		finish:             ctrl.Finish,
	}

	cli := &cli{
		commands: cmds,
		name:     "blar",
		usage:    mocks.usage,
		version:  mocks.version,

		flagsFromEnv: mocks.flagsFromEnv,
		cmdFlagsFromEnv: map[string]flags.FromEnv{
			"foo": mocks.cmdFooFlagsFromEnv,
			"bar": mocks.cmdBarFlagsFromEnv,
		},

		os: mocks.os,
	}

	return cli, mocks
}

func TestCLI(t *testing.T) {
	for _, tc := range []struct {
		args               [][]string
		cmdType            cmdType
		cliBarFlagValue    bool
		cmdBarFlagValue    bool
		fillFlagsCalled    bool
		cmdFillFlagsCalled bool
		usageCmdCalled     bool
		usageGlobalCalled  bool
		versionCalled      bool
		runnerCalled       bool
		cmdErrMessage      string
		errCode            command.CmdErrCode
		err                string
	}{
		//
		// SINGLE COMMAND TESTS
		//
		// bad argument tests
		{
			args:           [][]string{{"-bogus"}},
			cmdType:        singleCmd,
			usageCmdCalled: true,
			errCode:        command.CmdErrCodeBadInput,
			err:            "foo: flag provided but not defined: -bogus\n\n",
		},
		// -help tests
		{
			args:               [][]string{{"-help"}, {"-h"}},
			cmdType:            singleCmd,
			cmdFillFlagsCalled: true,
			usageCmdCalled:     true,
			errCode:            command.CmdErrCodeNoError,
		},
		// -version tests
		{
			args:               [][]string{{"-version"}, {"-v"}},
			cmdType:            singleCmd,
			cmdFillFlagsCalled: true,
			versionCalled:      true,
			errCode:            command.CmdErrCodeNoError,
		},
		// cmd success
		{
			args:               [][]string{{"-bar", "baz"}},
			cmdType:            singleCmd,
			cmdBarFlagValue:    true,
			cmdFillFlagsCalled: true,
			runnerCalled:       true,
			errCode:            command.CmdErrCodeNoError,
		},
		// cmd err
		{
			args:               [][]string{{"-bar", "baz"}},
			cmdType:            singleCmd,
			cmdBarFlagValue:    true,
			cmdFillFlagsCalled: true,
			runnerCalled:       true,
			errCode:            command.CmdErrCodeError,
			cmdErrMessage:      "Gah!",
			err:                "foo: Gah!\n\n",
		},
		//
		// MULTIPLE COMMAND TESTS
		//
		// bad argument tests
		{
			args:              [][]string{nil, {}},
			cmdType:           multipleCmds,
			fillFlagsCalled:   true,
			usageGlobalCalled: true,
			errCode:           command.CmdErrCodeBadInput,
			err:               "no command specified\n\n",
		},
		{
			args:              [][]string{{"bogus"}, {"bogus", "-version"}},
			cmdType:           multipleCmds,
			fillFlagsCalled:   true,
			usageGlobalCalled: true,
			errCode:           command.CmdErrCodeBadInput,
			err:               "unknown command: \"bogus\"\n\n",
		},
		{
			args:              [][]string{{"-bogus"}},
			cmdType:           multipleCmds,
			usageGlobalCalled: true,
			errCode:           command.CmdErrCodeBadInput,
			err:               "flag provided but not defined: -bogus\n\n",
		},
		{
			args:            [][]string{{"foo", "-bogus"}},
			cmdType:         multipleCmds,
			fillFlagsCalled: true,
			usageCmdCalled:  true,
			errCode:         command.CmdErrCodeBadInput,
			err:             "foo: flag provided but not defined: -bogus\n\n",
		},
		// -help tests
		{
			args: [][]string{
				{"help"},
				{"-help"},
				{"-h"},
				{"help", "bogus"},
				{"-help", "bogus"},
				{"-h", "bogus"},
				{"bogus", "-help"},
				{"bogus", "-h"},
			},
			cmdType:           multipleCmds,
			fillFlagsCalled:   true,
			usageGlobalCalled: true,
			errCode:           command.CmdErrCodeNoError,
		},
		{
			args: [][]string{
				{"help", "foo"},
				{"-help", "foo"},
				{"-h", "foo"},
				{"foo", "-help"},
				{"foo", "-h"},
			},
			cmdType:            multipleCmds,
			fillFlagsCalled:    true,
			cmdFillFlagsCalled: true,
			usageCmdCalled:     true,
			errCode:            command.CmdErrCodeNoError,
		},
		// -version tests
		{
			args: [][]string{
				{"version"},
				{"-version"},
				{"-v"},
				{"version", "foo"},
				{"-version", "foo"},
				{"-v", "foo"},
				{"version", "bogus"},
				{"-version", "bogus"},
				{"-v", "bogus"},
			},
			cmdType:         multipleCmds,
			fillFlagsCalled: true,
			versionCalled:   true,
			errCode:         command.CmdErrCodeNoError,
		},
		{
			args: [][]string{
				{"foo", "-version"},
				{"foo", "-v"},
			},
			cmdType:            multipleCmds,
			fillFlagsCalled:    true,
			cmdFillFlagsCalled: true,
			versionCalled:      true,
			errCode:            command.CmdErrCodeNoError,
		},
		// command success
		{
			args:               [][]string{{"-bar", "foo", "-bar", "baz"}},
			cmdType:            multipleCmds,
			cliBarFlagValue:    true,
			cmdBarFlagValue:    true,
			fillFlagsCalled:    true,
			cmdFillFlagsCalled: true,
			runnerCalled:       true,
			errCode:            command.CmdErrCodeNoError,
		},
		// command err
		{
			args:               [][]string{{"-bar", "foo", "-bar", "baz"}},
			cmdType:            multipleCmds,
			cliBarFlagValue:    true,
			cmdBarFlagValue:    true,
			cmdFillFlagsCalled: true,
			fillFlagsCalled:    true,
			runnerCalled:       true,
			errCode:            command.CmdErrCodeError,
			cmdErrMessage:      "Gah!",
			err:                "foo: Gah!\n\n",
		},
	} {
		for _, args := range tc.args {
			assert.Group(
				fmt.Sprintf(
					`TestCLI("%s", %s, %d)`,
					strings.Join(args, " "),
					tc.cmdType,
					int(tc.errCode),
				),
				t,
				func(g *assert.G) {
					c, mocks := newCLIAndMocks(g, tc.cmdType)
					defer mocks.finish()

					cmdBarFlag := false
					cliBarFlag := false

					fooCmd := c.command("foo")

					c.flags.BoolVar(&cliBarFlag, "bar", false, "")
					fooCmd.Flags.BoolVar(&cmdBarFlag, "bar", false, "")

					mocks.os.EXPECT().Args().Return(append([]string{c.name}, args...))

					if tc.fillFlagsCalled {
						mocks.flagsFromEnv.EXPECT().Fill()
					}

					if tc.cmdFillFlagsCalled {
						mocks.cmdFooFlagsFromEnv.EXPECT().Fill()
					}

					if tc.usageCmdCalled {
						mocks.usage.EXPECT().Command(
							fooCmd,
							mocks.flagsFromEnv,
							mocks.cmdFooFlagsFromEnv,
						)
					}

					if tc.usageGlobalCalled {
						mocks.usage.EXPECT().Global(c.commands, mocks.flagsFromEnv)
					}

					if tc.versionCalled {
						mocks.version.EXPECT().Print()
					}

					if tc.runnerCalled {
						cmdErr := command.NoError()
						if tc.cmdErrMessage != "" {
							cmdErr = fooCmd.Error(tc.cmdErrMessage)
						}
						mocks.fooRunner.EXPECT().Run(fooCmd, []string{"baz"}).Return(cmdErr)
					}

					if tc.err != "" {
						mocks.os.EXPECT().Stderr().Return(mocks.stderr)
					}

					mocks.os.EXPECT().Exit(int(tc.errCode))

					c.Main()

					assert.Equal(g, tc.cliBarFlagValue, cliBarFlag)
					assert.Equal(g, tc.cmdBarFlagValue, cmdBarFlag)
					assert.Equal(g, mocks.stderr.String(), tc.err)
				},
			)
		}
	}
}
