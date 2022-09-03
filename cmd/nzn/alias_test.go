//
// nazuna/cmd/nzn :: alias_test.go
//
//   Copyright (c) 2013-2022 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"testing"

	"github.com/hattya/go.cli"
)

func TestAlias(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
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
			cmd: []string{"mkdir", ".nzn/r/a/.config/gocode"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.config/gocode/config.json"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.gitconfig"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.vimrc"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/a/.vim/syntax"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.vim/syntax/vim.vim"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b/1"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/b/1/.vim/syntax"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/1/.vim/syntax/go.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "layer", "b/1"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .config/ --> a
				link .gitconfig --> a
				link .vim/syntax/go.vim --> b/1
				link .vim/syntax/vim.vim --> a
				link .vimrc --> a
				5 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b/2"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/b/2/vimfiles/syntax"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/2/vimfiles/syntax/go.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/2", ".vim", "vimfiles"},
		},
		{
			cmd: []string{"export", "APPDATA=$wc/AppData/Roaming"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/2", ".config/gocode/config.json", "$APPDATA/gocode/config.json"},
		},
		{
			cmd: []string{"nzn", "layer", "b/2"},
		},
		{
			cmd: []string{"mkdir", "AppData/Roaming"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink .config/ -/- a
				unlink .vim/syntax/go.vim -/- b/1
				unlink .vim/syntax/vim.vim -/- a
				link AppData/Roaming/gocode/ --> a:.config/gocode/
				link vimfiles/syntax/go.vim --> b/2
				link vimfiles/syntax/vim.vim --> a:.vim/syntax/vim.vim
				3 updated, 3 removed, 0 failed
			`),
		},
		{
			cmd: []string{"rm", "-r", "AppData/Roaming"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link AppData/Roaming/ --> a:.config/
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.curlrc"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/2", ".curlrc", "$APPDATA/_curlrc"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink AppData/Roaming/ -/- a:.config/
				link AppData/Roaming/_curlrc --> a:.curlrc
				link AppData/Roaming/gocode/ --> a:.config/gocode/
				2 updated, 1 removed, 0 failed
			`),
		},
	}
	if err := s.exec(t); err != nil {
		t.Error(err)
	}
}

func TestAliasError(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "alias"},
			out: cli.Dedent(`
				nzn: no repository found in '.+' \(\.nzn not found\)! (re)
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"touch", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "alias"},
			out: cli.Dedent(`
				nzn: ` + path(".nzn/state.json") + `: unexpected end of JSON input
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "alias"},
			out: cli.Dedent(`
				nzn alias: --layer flag is required
				usage: nzn alias -l <layer> <src> <dst>

				create an alias for the specified path

				  Change the location of <src> to <dst>. <src> should be existed in the lower layer
				  than <dst>, and <src> is treated as <dst> in the layer <layer>. If <src> does
				  not match any locations on update, it will be ignored without error.

				  You can refer environment variables in <dst>. Supported formats are ${var}
				  and $var.

				options:

				  -l, --layer <layer>    layer name

				[2]
			`),
		},
		{
			cmd: []string{"nzn", "alias", "-l", "a", "src", "dst"},
			out: cli.Dedent(`
				nzn: layer 'a' does not exist!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "a"},
			out: cli.Dedent(`
				nzn: invalid arguments
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b/1"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b", "src", "dst"},
			out: cli.Dedent(`
				nzn: layer 'b' is abstract
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "../src", "dst"},
			out: cli.Dedent(`
				nzn: '../src' is not under root
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "../dst"},
			out: cli.Dedent(`
				nzn: '../dst' is not under root
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "src"},
			out: cli.Dedent(`
				nzn: 'src' and 'src' are the same path
				[1]
			`),
		},
		{
			cmd: []string{"touch", ".nzn/r/b/1/dst"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "dst"},
			out: cli.Dedent(`
				nzn: 'dst' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-qf", "b/1/dst"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "dst"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "dst"},
			out: cli.Dedent(`
				nzn: alias 'dst' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"export", "ROOT=.."},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/src"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b/2"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/2", "src", "$ROOT/dst"},
		},
		{
			cmd: []string{"nzn", "layer", "b/2"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				nzn: '` + path("../dst") + `' is not under root
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "b/1"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/1"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "c/1", "_", "src"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/2"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "c/2", "-p", ".", "_", "src"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/3"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "c/3", "-a", "_", "src"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "d"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "d", "src", "$ROOT/dst"},
		},
		{
			cmd: []string{"touch", "_"},
		},
		{
			cmd: []string{"nzn", "layer", "c/1"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				nzn: link '` + path("../dst") + `' is not under root
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "c/2"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				nzn: link '` + path("../dst") + `' is not under root
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "c/3"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				nzn: subrepo '` + path("../dst") + `' is not under root
				[1]
			`),
		},
		{
			cmd: []string{"rm", "_"},
		},
	}
	if err := s.exec(t); err != nil {
		t.Error(err)
	}
}
