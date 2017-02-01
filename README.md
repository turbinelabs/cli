
[//]: # ( Copyright 2017 Turbine Labs, Inc.                                   )
[//]: # ( you may not use this file except in compliance with the License.    )
[//]: # ( You may obtain a copy of the License at                             )
[//]: # (                                                                     )
[//]: # (     http://www.apache.org/licenses/LICENSE-2.0                      )
[//]: # (                                                                     )
[//]: # ( Unless required by applicable law or agreed to in writing, software )
[//]: # ( distributed under the License is distributed on an "AS IS" BASIS,   )
[//]: # ( WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or     )
[//]: # ( implied. See the License for the specific language governing        )
[//]: # ( permissions and limitations under the License.                      )

# turbinelabs/cli

[![Apache 2.0](https://img.shields.io/hexpm/l/plug.svg)](LICENSE)
[![GoDoc](https://https://godoc.org/github.com/turbinelabs/cli?status.svg)](https://https://godoc.org/github.com/turbinelabs/cli)
[![CircleCI](https://circleci.com/gh/turbinelabs/cli.svg?style=shield)](https://circleci.com/gh/turbinelabs/cli)

The `cli` package provides a simple library for creating command-line
applications with multiple sub-commands.

## Motivation

We feel strongly that all Turbine Labs executables should have high-quality
and consistent ergonomics. We use the `cli` package for all our command line
tools, private and public.

Why on earth did we choose to roll our own CLI package, with several high quality alternatives already
available (eg [github.com/urfave/cli](https://github.com/urfave/cli), [github.com/spf13/cobra](https://github.com/spf13/cobra))? As with most libraries, it came
out of a combination of good intentions and hubris.

Chiefly, we wanted:

  - something fairly lightweight
  - to use the [existing Go flag package](https://golang.org/pkg/flag/)
  - to support both single- and sub-command apps
  - to minimize external dependencies
  - for flags to be configurable via environment flag
  - for usage text to be both auto-generated and fairly customizable

We started with the excellent [github.com/jonboulle/yacli](https://github.com/jonboulle/yacli), a "micro-pseudo-framework" which recognizes a central truth:

> All CLI frameworks suck. This one just sucks slightly less.

Indeed we were attracted to its simplicity and ease of use, and used it for a
time, but quickly grew weary of the attendant boilerplate necessary for each
new project. We began to adapt it, to centralize code, and ended up writing a
whole thing.

Now that we have it, we like it, and we welcome you to use it, or not, write
your own, or contribute to ours, or whatever. It's all fine with us.

## Requirements

- Go 1.7.4 or later (previous versions may work, but we don't build or test against them)

## Dependencies

The cli project depends on our
[nonstdlib package](https://github.com/turbinelabs/nonstdlib); The tests depend
on our [test package](https://github.com/turbinelabs/test) and
[gomock](https://github.com/golang/mock). It should always be safe to use HEAD
of all master branches of Turbine Labs open source projects together, or to
vendor them with the same git tag.

Additionally, we vendor
[github.com/mattn/go-isatty](https://github.com/mattn/go-isatty) and
[golang.org/x/sys/unix](https://godoc.org/golang.org/x/sys/unix).
This should be considered an opaque implementation detail,
see [Vendoring](http://github.com/turbinelabs/developer/blob/master/README.md#vendoring) for more detail.

## Install

```
go get -u github.com/turbinelabs/cli/...
```

## Clone/Test

```
mkdir -p $GOPATH/src/turbinelabs
git clone https://github.com/turbinelabs/cli.git > $GOPATH/src/turbinelabs/cli
go test github.com/turbinelabs/cli/...
```

## Godoc

[`cli`](https://godoc.org/github.com/turbinelabs/cli)

## Features

The `cli` package includes:

  - Support for both global and per-sub-command flags
  - Automatic generation of help and version flags and (if appropriate)
    sub-commands
  - Basic terminal-aware formatting of usage text, including pre-formatted blocks,
    and bold and underlined text
  - Auto-wrapping of usage text to the terminal width
  - Support for the use of environment variables to set both global and
    per-sub-command flags

#### Environment Variables

Both global and per sub-command flags can be set by environment variable.
Vars are all-caps, underscore-delimited, and replace all non-alphanumeric
symbols with underscores. They are prefixed by the executable name, and
by then by the subcommand if relevant.

For example, the following are equivalent:

    somecmd --global-flag=a somesubcmd --cmd-flag=b
    SOMECMD_GLOBAL_FLAG=a SOMECMD_SOMESUBCMD_CMD_FLAG=b somecmd

A runtime validation is available to ensure that there are no variable name
collisions for a given CLI.

#### Help Text

Help text is generated from:

  - The description passed into [`cli.NewWithSubCmds`](https://godoc.org/github.com/turbinelabs/cli/#NewWithSubCmds)
  - The `.Description`, `.Summary` and `.Usage` fields of each
  [`command.Cmd`](https://godoc.org/github.com/turbinelabs/cli/command/#Cmd)
  - The `.Usage` field of each [`flag.Flag`](https://golang.org/pkg/flag/#Flag)

Since our target is Terminal windows, we chose to keep formatting fairly simple.

All text is passed through [text/template](https://golang.org/pkg/text/template);
As such, double curly braces ("{{") in text will trigger template actions.

The cli package provides two text-styling functions for help text, for terminals
that support them:

```go
`text may rendered in {{bold "some bold text"}}`
`text may be {{ul "some underlined text}}`
```

To render text that contains "{{", use a pipeline action:

```go
`here are some curlies {{ "{{something inside braces}}" }}`
```

Help text is also passed through [go/doc](https://golang.org/pkg/go/doc), to
support a few simple formatting primitives:

- Text wraps to the current terminal width, if available, or 80 columns
  otherwise.
- Whitespace, including single newlines, is collapsed in a given paragraph.
- New paragraphs are signaled with two newlines.
- A 4-space indent signals pre-formatted text, which is not wrapped.

## Examples

[Single-](https://godoc.org/github.com/turbinelabs/cli/#example__singleCommand)
and [multiple-command](https://godoc.org/github.com/turbinelabs/cli/#example__subCommands)
examples are available in the [godoc](https://godoc.org/github.com/turbinelabs/cli).

## Versioning

Please see [Versioning of Turbine Labs Open Source Projects](http://github.com/turbinelabs/developer/blob/master/README.md#versioning).

## Pull Requests

Patches accepted! Please see [Contributing to Turbine Labs Open Source Projects](http://github.com/turbinelabs/developer/blob/master/README.md#contributing).

## Code of Conduct

All Turbine Labs open-sourced projects are released with a
[Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in our
projects you agree to abide by its terms, which will be vigorously enforced.
