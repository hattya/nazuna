//
// nazuna/cmd/nzn :: link_test.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"os"
	"testing"

	"github.com/hattya/go.cli"
)

func TestLink(t *testing.T) {
	sep := string(os.PathListSeparator)
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"mkdir", "$public/go/misc/vim"},
		},
		{
			cmd: []string{"mkdir", "$public/gocode/src/github.com/nsf/gocode/vim"},
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.vimrc"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "$public/go/misc/vim", ".vim/bundle/golang"},
		},
		{
			cmd: []string{"export", "GOPATH="},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "-p", "$GOPATH" + sep + "$public/gocode", "src/github.com/nsf/gocode/vim", ".vim/bundle/gocode"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vim/bundle/gocode/ --> .+` + quote("/gocode/src/github.com/nsf/gocode/vim/") + ` (re)
				link .vim/bundle/golang/ --> .+` + quote("/go/misc/vim/") + ` (re)
				link .vimrc --> a
				3 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"rm", "-r", "$public/go"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink .vim/bundle/golang/ -/- .+` + quote("/go/misc/vim/") + ` (re)
				0 updated, 1 removed, 0 failed
			`),
		},
		{
			cmd: []string{"rm", "-r", "$public"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink .vim/bundle/gocode/ -/- .+` + quote("/gocode/src/github.com/nsf/gocode/vim/") + ` (re)
				0 updated, 1 removed, 0 failed
			`),
		},
		{
			cmd: []string{"ls", "."},
			out: cli.Dedent(`
				.nzn/
				.vimrc
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestLinkError(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"nzn", "link"},
			out: cli.Dedent(`
				nzn: no repository found in '.+' \(\.nzn not found\)! (re)
				[1]
			`),
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"touch", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "link"},
			out: cli.Dedent(`
				nzn: ` + path(".nzn/state.json") + `: unexpected end of JSON input
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "link"},
			out: cli.Dedent(`
				nzn link: --layer flag is required
				usage: nzn link -l <layer> [-p <path>] <src> <dst>

				create a link for the specified path

				  link is used to create a link of <src> to <dst>, and will be managed by
				  update. If <src> is not found on update, it will be ignored without error.

				  The value of --path flag is a list of directories like PATH or GOPATH
				  environment variables, and it is used to search <src>.

				  You can refer environment variables in <path> and <src>. Supported formats
				  are ${var} and $var.

				options:

				  -l, --layer <layer>    layer name
				  -p, --path <path>      list of directories to search <src>

				[2]
			`),
		},
		{
			cmd: []string{"nzn", "link", "-l", "a"},
			out: cli.Dedent(`
				nzn: invalid arguments
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
			out: cli.Dedent(`
				nzn: layer 'a' does not exist!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "../dst"},
			out: cli.Dedent(`
				nzn: '../dst' is not under root
				[1]
			`),
		},
		{
			cmd: []string{"touch", ".nzn/r/a/dst"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
			out: cli.Dedent(`
				nzn: 'dst' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-qf", "a/dst"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/a/dst"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/dst/1"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
			out: cli.Dedent(`
				nzn: 'dst' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-qrf", "a/dst"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
			out: cli.Dedent(`
				nzn: link 'dst' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"touch", "src"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link dst --> src
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "b", "src", "dst"},
		},
		{
			cmd: []string{"rm", "dst"},
		},
		{
			cmd: []string{"touch", "_"},
		},
		{
			cmd: []string{"ln", "-s", "_", "dst"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink dst -/- src
				nzn: not linked to 'src'
				[1]
			`),
		},
		{
			cmd: []string{"rm", "dst"},
		},
		{
			cmd: []string{"rm", "_"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/dst"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				warning: link: 'dst' exists in the repository
				link dst --> b
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/1"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "c", "src", "dst"},
			out: cli.Dedent(`
				nzn: layer 'c' is abstract
				[1]
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}
