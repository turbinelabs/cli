/*
Copyright 2017 Turbine Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	tbnflag "github.com/turbinelabs/nonstdlib/flag"
	"github.com/turbinelabs/nonstdlib/flag/usage"
	tbnos "github.com/turbinelabs/nonstdlib/os"
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

	assert.DeepEqual(t, fooCli.Validate(ValidateSkipHelpText), wantErr)
}

func TestValidateHelpText(t *testing.T) {
	barCmd := &command.Cmd{Name: "bar"}
	barBazCmd := &command.Cmd{Name: "bar-baz"}
	fooCli := mkNew(app.App{Name: "foo"}, barCmd, barBazCmd)

	assert.Nil(t, fooCli.Validate())

	barCmd = &command.Cmd{Name: "bar", Description: "this won't {{work}}"}
	barBazCmd = &command.Cmd{Name: "bar-baz"}
	fooCli = mkNew(app.App{Name: "foo"}, barCmd, barBazCmd)

	assert.ErrorContains(t, fooCli.Validate(), `function "work" not defined`)

	barCmd = &command.Cmd{Name: "bar"}
	barBazCmd = &command.Cmd{Name: "bar-baz", Description: "this won't work {{either}}"}
	fooCli = mkNew(app.App{Name: "foo"}, barCmd, barBazCmd)

	assert.ErrorContains(t, fooCli.Validate(), `function "either" not defined`)

	// ignores help text in this case
	assert.Nil(t, fooCli.Validate(ValidateSkipHelpText))
}

type cmdType string

const (
	noSubCmd     cmdType = "no-sub-cmd"
	multipleCmds         = "multiple-cmds"
)

type cliTestMocks struct {
	t                  testing.TB
	fooRunner          *command.MockRunner
	barRunner          *command.MockRunner
	usage              *app.MockUsage
	version            *app.MockVersion
	flagsFromEnv       *tbnflag.MockFromEnv
	cmdFooFlagsFromEnv *tbnflag.MockFromEnv
	cmdBarFlagsFromEnv *tbnflag.MockFromEnv
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
		flagsFromEnv:       tbnflag.NewMockFromEnv(ctrl),
		cmdFooFlagsFromEnv: tbnflag.NewMockFromEnv(ctrl),
		cmdBarFlagsFromEnv: tbnflag.NewMockFromEnv(ctrl),
		os:                 tbnos.NewMockOS(ctrl),
		stderr:             &bytes.Buffer{},
		finish:             ctrl.Finish,
	}

	var fromEnvMap map[string]tbnflag.FromEnv

	if cType == multipleCmds {
		fromEnvMap = map[string]tbnflag.FromEnv{
			"foo": mocks.cmdFooFlagsFromEnv,
			"bar": mocks.cmdBarFlagsFromEnv,
		}
	} else {
		fromEnvMap = map[string]tbnflag.FromEnv{
			"blar": mocks.cmdFooFlagsFromEnv,
		}
	}

	cli := &cli{
		commands: cmds,
		name:     "blar",
		app: app.App{
			HasSubCmds: cType != noSubCmd,
		},
		usage:   mocks.usage,
		version: mocks.version,

		flagsFromEnv:    mocks.flagsFromEnv,
		cmdFlagsFromEnv: fromEnvMap,

		os: mocks.os,
	}

	return cli, mocks
}

func TestCLI(t *testing.T) {
	for _, tc := range []struct {
		args               [][]string
		cmdType            cmdType
		cliBarFlagValue    string
		cmdBarFlagValue    string
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
		// NO SUB COMMAND TESTS
		//
		// bad argument tests
		{
			args:           [][]string{{"-bogus"}},
			cmdType:        noSubCmd,
			usageCmdCalled: true,
			errCode:        command.CmdErrCodeBadInput,
			err:            "foo: flag provided but not defined: -bogus\n\n",
		},
		// required arg test
		{
			args:               [][]string{{"baz"}},
			cmdType:            noSubCmd,
			cmdFillFlagsCalled: true,
			usageCmdCalled:     true,
			errCode:            command.CmdErrCodeBadInput,
			err:                "foo: \n  --bar is a required flag\n\n",
		},
		// -help tests
		{
			args:               [][]string{{"-help"}, {"-h"}},
			cmdType:            noSubCmd,
			cmdFillFlagsCalled: true,
			usageCmdCalled:     true,
			errCode:            command.CmdErrCodeNoError,
		},
		// -version tests
		{
			args:               [][]string{{"-version"}, {"-v"}},
			cmdType:            noSubCmd,
			cmdFillFlagsCalled: true,
			versionCalled:      true,
			errCode:            command.CmdErrCodeNoError,
		},
		// cmd success
		{
			args:               [][]string{{"-bar", "a", "baz"}},
			cmdType:            noSubCmd,
			cmdBarFlagValue:    "a",
			cmdFillFlagsCalled: true,
			runnerCalled:       true,
			errCode:            command.CmdErrCodeNoError,
		},
		// cmd err
		{
			args:               [][]string{{"-bar", "a", "baz"}},
			cmdType:            noSubCmd,
			cmdBarFlagValue:    "a",
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
			err:               "no command specified\n--bar is a required global flag\n\n",
		},
		{
			args:              [][]string{{"bogus"}, {"bogus", "-version"}},
			cmdType:           multipleCmds,
			fillFlagsCalled:   true,
			usageGlobalCalled: true,
			errCode:           command.CmdErrCodeBadInput,
			err:               "unknown command: \"bogus\"\n--bar is a required global flag\n\n",
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
		// missing flag test
		{
			args:               [][]string{{"foo", "baz"}},
			cmdType:            multipleCmds,
			fillFlagsCalled:    true,
			cmdFillFlagsCalled: true,
			usageCmdCalled:     true,
			errCode:            command.CmdErrCodeBadInput,
			err:                "foo: \n  --bar is a required global flag\n  --bar is a required flag\n\n",
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
			args:               [][]string{{"-bar", "a", "foo", "-bar", "b", "baz"}},
			cmdType:            multipleCmds,
			cliBarFlagValue:    "a",
			cmdBarFlagValue:    "b",
			fillFlagsCalled:    true,
			cmdFillFlagsCalled: true,
			runnerCalled:       true,
			errCode:            command.CmdErrCodeNoError,
		},
		// command err
		{
			args:               [][]string{{"-bar", "a", "foo", "-bar", "b", "baz"}},
			cmdType:            multipleCmds,
			cliBarFlagValue:    "a",
			cmdBarFlagValue:    "b",
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

					var cmdBarFlag, cliBarFlag string

					fooCmd := c.command("foo")

					u := usage.New("").SetRequired().SetSensitive().String()
					c.flags.StringVar(&cliBarFlag, "bar", "", u)
					fooCmd.Flags.StringVar(&cmdBarFlag, "bar", "", u)

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
						mocks.version.EXPECT().Describe().Return("version")
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
