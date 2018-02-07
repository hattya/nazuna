//
// nazuna/cmd/nzn :: help_test.go
//
//   Copyright (c) 2013-2018 Akinori Hattori <hattya@gmail.com>
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
	"testing"

	"github.com/hattya/go.cli"
)

var (
	helpUsage = cli.Dedent(`
		usage: nzn help [<command>]

		show help for a specified command
	`)
	helpOut = cli.Dedent(`
		Nazuna - A layered dotfiles management

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
	`)
)

func TestHelp(t *testing.T) {
	s := script{
		{
			cmd: []string{"nzn", "--help"},
			out: helpOut + "\n",
		},
		{
			cmd: []string{"nzn", "help"},
			out: helpOut + "\n",
		},
		{
			cmd: []string{"nzn"},
			out: fmt.Sprintf(cli.Dedent(`
				%v
				[1]
			`), helpOut),
		},
		{
			cmd: []string{"nzn", "--nazuna"},
			out: fmt.Sprintf(cli.Dedent(`
				nzn: flag provided but not defined: -nazuna
				%v
				[2]
			`), helpOut),
		},
		{
			cmd: []string{"nzn", "nazuna"},
			out: fmt.Sprintf(cli.Dedent(`
				nzn: unknown command 'nazuna'
				%v
				[1]
			`), helpOut),
		},
		{
			cmd: []string{"nzn", "help", "help"},
			out: helpUsage + "\n",
		},
		{
			cmd: []string{"nzn", "help", "--help"},
			out: helpUsage + "\n",
		},
		{
			cmd: []string{"nzn", "help", "--nazuna"},
			out: fmt.Sprintf(cli.Dedent(`
				nzn help: flag provided but not defined: -nazuna
				%v
				[2]
			`), helpUsage),
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
			out: cli.Dedent(`
				nzn: unknown command 'nazuna'
				type 'nzn help' for usage
				[1]
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}
