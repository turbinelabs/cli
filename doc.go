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
