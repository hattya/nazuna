//
// nazuna/cmd/nzn :: init_test.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"testing"

	"github.com/hattya/go.cli"
)

func TestInit(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git", "$wc"},
		},
		{
			cmd: []string{"ls", "$wc/.nzn"},
			out: cli.Dedent(`
				r/
			`),
		},
		{
			cmd: []string{"ls", "$wc/.nzn/r"},
			out: cli.Dedent(`
				.git/
				nazuna.json
			`),
		},
		{
			cmd: []string{"cat", "$wc/.nzn/r/nazuna.json"},
			out: cli.Dedent(`
				[]
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestInitError(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"nzn", "init", "$wc"},
			out: cli.Dedent(`
				nzn init: --vcs flag is required
				usage: nzn init --vcs <type> [<path>]

				create a new repository in the specified directory

				  Create a new repository in <path>. If <path> does not exist, it will be
				  created.

				  If <path> is not specified, the current working diretory is used.

				options:

				  --vcs <type>    vcs type

				[2]
			`),
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "cvs", "$wc"},
			out: cli.Dedent(`
				nzn: unknown vcs 'cvs'
				[1]
			`),
		},
		{
			cmd: []string{"mkdir", "$wc/.nzn/r"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git", "$wc"},
			out: cli.Dedent(`
				nzn: repository '.+' already exists! (re)
				[1]
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}
