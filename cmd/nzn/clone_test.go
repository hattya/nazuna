//
// nazuna/cmd/nzn :: clone_test.go
//
//   Copyright (c) 2013-2024 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"testing"

	"github.com/hattya/go.cli"
)

func TestClone(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"git", "init", "-q", "$public/repo"},
		},
		{
			cmd: []string{"cd", "$public/repo"},
		},
		{
			cmd: []string{"touch", "nazuna.json"},
		},
		{
			cmd: []string{"git", "add", "."},
		},
		{
			cmd: []string{"git", "commit", "-qm", "."},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"nzn", "clone", "--vcs", "git", "$public/repo", "wc"},
			out: cli.Dedent(`
				Cloning into '` + path("wc/.nzn/r") + `'...
				done.
			`),
		},
		{
			cmd: []string{"ls", "wc/.nzn"},
			out: cli.Dedent(`
				r/
			`),
		},
		{
			cmd: []string{"ls", "wc/.nzn/r"},
			out: cli.Dedent(`
				.git/
				nazuna.json
			`),
		},
	}
	if err := s.exec(t); err != nil {
		t.Error(err)
	}
}

func TestCloneError(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"nzn", "clone"},
			out: cli.Dedent(`
				nzn: invalid arguments
				[1]
			`),
		},
		{
			cmd: []string{"git", "init", "-q", "$public/repo"},
		},
		{
			cmd: []string{"nzn", "clone", "$public/repo", "wc"},
			out: cli.Dedent(`
				nzn clone: --vcs flag is required (re)
				usage: nzn clone --vcs <type> <repository> [<path>]

				create a copy of an existing repository

				  Create a copy of an existing repository in <path>. If <path> does not exist,
				  it will be created.

				  If <path> is not specified, the current working directory is used.

				options:

				  --vcs <type>    vcs type

				[2]
			`),
		},
		{
			cmd: []string{"nzn", "clone", "--vcs", "cvs", "$public/repo", "wc"},
			out: cli.Dedent(`
				nzn: unknown vcs 'cvs'
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git", "wc"},
		},
		{
			cmd: []string{"nzn", "clone", "--vcs", "git", "$public/repo", "wc"},
			out: cli.Dedent(`
				nzn: repository 'wc' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "clone", "--vcs", "git", "$public/repo"},
			out: cli.Dedent(`
				nzn: repository '.' already exists!
				[1]
			`),
		},
	}
	if err := s.exec(t); err != nil {
		t.Error(err)
	}
}
