/*
Copyright 2018 Turbine Labs, Inc.

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

package terminal

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"

	tbnos "github.com/turbinelabs/nonstdlib/os"
	"github.com/turbinelabs/test/assert"
	"github.com/turbinelabs/test/io"
)

func TestAsk(t *testing.T) {
	ctrl := gomock.NewController(assert.Tracing(t))
	defer ctrl.Finish()

	in := bytes.NewBufferString("yes\neh?\n")
	out := bytes.NewBuffer(make([]byte, 0, 128))

	os := tbnos.NewMockOS(ctrl)
	os.EXPECT().Stdin().Return(in).AnyTimes()
	os.EXPECT().Stdout().Return(out).AnyTimes()

	answer, err := Ask(os, "are you %s?", "ready")
	assert.Nil(t, err)
	assert.True(t, answer)
	assert.Equal(t, out.String(), "are you ready? [y/N]: ")
	out.Reset()

	answer, err = Ask(os, "are you %s?", "sure")
	assert.Nil(t, err)
	assert.False(t, answer)
	assert.Equal(t, out.String(), "are you sure? [y/N]: ")
}

func TestAskError(t *testing.T) {
	ctrl := gomock.NewController(assert.Tracing(t))
	defer ctrl.Finish()

	os := tbnos.NewMockOS(ctrl)
	os.EXPECT().Stdin().Return(io.NewFailingReader())
	os.EXPECT().Stdout().Return(bytes.NewBuffer(make([]byte, 0, 128)))

	answer, err := Ask(os, "are you ready?")
	assert.NonNil(t, err)
	assert.False(t, answer)
}
