package terminal

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	tbnos "github.com/turbinelabs/nonstdlib/os"
	"github.com/turbinelabs/test/assert"
	testio "github.com/turbinelabs/test/io"
)

func TestPrompt(t *testing.T) {
	ctrl := gomock.NewController(assert.Tracing(t))
	defer ctrl.Finish()

	in := bytes.NewBufferString("response\n")
	out := bytes.NewBuffer(make([]byte, 0, 128))

	os := tbnos.NewMockOS(ctrl)
	os.EXPECT().Stdin().Return(in)
	os.EXPECT().Stdout().Return(out)

	resp, err := Prompt(os, "prompt %d", 1)
	assert.Nil(t, err)
	assert.Equal(t, resp, "response")
	assert.Equal(t, out.String(), "prompt 1")

	in = bytes.NewBufferString("response\r\r\r\n")
	out.Reset()
	os.EXPECT().Stdin().Return(in)
	os.EXPECT().Stdout().Return(out)

	resp, err = Prompt(os, "prompt %d", 2)
	assert.Nil(t, err)
	assert.Equal(t, resp, "response")
	assert.Equal(t, out.String(), "prompt 2")
}

func TestPromptLongInput(t *testing.T) {
	ctrl := gomock.NewController(assert.Tracing(t))
	defer ctrl.Finish()

	in := bytes.NewBufferString(strings.Repeat("x", 300) + "\r\r\n")
	out := bytes.NewBuffer(make([]byte, 0, 128))

	os := tbnos.NewMockOS(ctrl)
	os.EXPECT().Stdin().Return(in)
	os.EXPECT().Stdout().Return(out)

	resp, err := Prompt(os, "prompt long %d", 1)
	assert.Nil(t, err)
	assert.Equal(t, resp, strings.Repeat("x", 300))
	assert.Equal(t, out.String(), "prompt long 1")
}

func TestPromptError(t *testing.T) {
	ctrl := gomock.NewController(assert.Tracing(t))
	defer ctrl.Finish()

	out := bytes.NewBuffer(make([]byte, 0, 128))

	os := tbnos.NewMockOS(ctrl)
	os.EXPECT().Stdin().Return(testio.NewFailingReader())
	os.EXPECT().Stdout().Return(out)

	resp, err := Prompt(os, "prompt err %d", 1)
	assert.NonNil(t, err)
	assert.Equal(t, resp, "")
	assert.Equal(t, out.String(), "prompt err 1")

	out.Reset()
	os.EXPECT().Stdin().Return(bytes.NewBufferString("x"))
	os.EXPECT().Stdout().Return(out)

	resp, err = Prompt(os, "prompt err %d", 2)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, resp, "x")
	assert.Equal(t, out.String(), "prompt err 2")

	out.Reset()
	os.EXPECT().Stdin().Return(bytes.NewBufferString(""))
	os.EXPECT().Stdout().Return(out)

	resp, err = Prompt(os, "prompt err %d", 3)
	assert.Equal(t, err, io.EOF)
	assert.Equal(t, resp, "")
	assert.Equal(t, out.String(), "prompt err 3")
}
