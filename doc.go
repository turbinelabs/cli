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

/*
The cli package provides a simple library for creating command-line
applications with multiple sub-commands. It supports both global and
per-subcommand flags, and automatically generates help and version
sub-commands.

Help Text

Help text is generated using go's text/template package. The
description passed to NewWithSubCmds, the Summary and Description
fields of command.Cmd struct, and the Usage field of go's flag.Flag
struct are used to create a template for the help text. As such,
double curly braces ("{{") in those fields trigger template actions.

The cli package provides two functions to aid in formatting help text,
shown here as example strings:

        `text may rendered in {{bold "boldface"}}`
        `text may be {{ul "underlined}}`

To render text that contains "{{", use a pipeline action:

        `here are some curlies {{ "{{tada}}" }}`
*/
package cli
