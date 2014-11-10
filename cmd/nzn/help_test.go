//
// nzn :: help_test.go
//
//   Copyright (c) 2013-2014 Akinori Hattori <hattya@gmail.com>
//
//   Permission is hereby granted, free of charge, to any person
//   obtaining a copy of this software and associated documentation files
//   (the "Software"), to deal in the Software without restriction,
//   including without limitation the rights to use, copy, modify, merge,
//   publish, distribute, sublicense, and/or sell copies of the Software,
//   and to permit persons to whom the Software is furnished to do so,
//   subject to the following conditions:
//
//   The above copyright notice and this permission notice shall be
//   included in all copies or substantial portions of the Software.
//
//   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//   EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
//   MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//   NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
//   BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
//   ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
//   CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//   SOFTWARE.
//

package main

import (
	"fmt"
	"strings"
	"testing"
)

const (
	helpUsage = `usage: nzn help [<command>]

show help for a specified command

`
	helpOut = `nazuna - A layered dotfiles management

commands:

  alias      create an alias for the specified path
  clone      create a copy of an existing repository
  help       show help for a specified command
  init       create a new repository in the specified directory
  layer      manage repository layers
  link       create a link for the specified path
  subrepo    manage subrepositories
  update     update working copy
  vcs        run the vcs command inside the repository
  version    show version information

options:

  -h, --help    show help
  --version     show version information

`
)

func TestHelp(t *testing.T) {
	usage := strings.TrimSuffix(helpUsage, "\n")
	out := strings.TrimSuffix(helpOut, "\n")
	s := script{
		{
			cmd: []string{"nzn", "--help"},
			out: helpOut,
		},
		{
			cmd: []string{"nzn", "help"},
			out: helpOut,
		},
		{
			cmd: []string{"nzn"},
			out: fmt.Sprintf(`%v
[1]
`, out),
		},
		{
			cmd: []string{"nzn", "--nazuna"},
			out: fmt.Sprintf(`nzn: flag .* not defined: -*nazuna (re)
%v
[2]
`, out),
		},
		{
			cmd: []string{"nzn", "nazuna"},
			out: fmt.Sprintf(`nzn: unknown command 'nazuna'
%v
[1]
`, out),
		},
		{
			cmd: []string{"nzn", "help", "help"},
			out: helpUsage,
		},
		{
			cmd: []string{"nzn", "help", "--help"},
			out: helpUsage,
		},
		{
			cmd: []string{"nzn", "help", "--nazuna"},
			out: fmt.Sprintf(`nzn help: flag .* not defined: -*nazuna (re)
%v
[2]
`, usage),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestHelpError(t *testing.T) {
	s := script{
		{
			cmd: []string{"nzn", "help", "nazuna"},
			out: `nzn: unknown command 'nazuna'
type 'nzn help' for usage
[1]
`,
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}
