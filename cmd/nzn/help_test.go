//
// nazuna/cmd/nzn :: help_test.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
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
