//
// nazuna/cmd/nzn :: update_test.go
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
	"testing"

	"github.com/hattya/go.cli"
)

func TestUpdate(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"cd", "w"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
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
			cmd: []string{"nzn", "vcs", "add", "a"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .gitconfig --> a
				link .vim/ --> a
				link .vimrc --> a
				3 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/.vimrc"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/b/.vim/syntax"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/.vim/syntax/go.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink .vim/ -/- a
				unlink .vimrc -/- a
				link .vim/syntax/go.vim --> b
				link .vim/syntax/vim.vim --> a
				link .vimrc --> b
				3 updated, 2 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/1"},
		},
		{
			cmd: []string{"touch", ".nzn/r/c/1/.screenrc"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/2"},
		},
		{
			cmd: []string{"touch", ".nzn/r/c/2/.tmux.conf"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "c"},
		},
		{
			cmd: []string{"nzn", "layer", "c/1"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .screenrc --> c/1
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "layer", "c/2"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink .screenrc -/- c/1
				link .tmux.conf --> c/2
				1 updated, 1 removed, 0 failed
			`),
		},
		{
			cmd: []string{"mkdir", ".nzn/r/b/.vim/autoload/go"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/.vim/autoload/go/complete.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vim/autoload/ --> b
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"rm", ".vim/autoload"},
		},
		{
			cmd: []string{"mkdir", ".vim/autoload"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vim/autoload/go/ --> b
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"rm", ".vim/autoload/go"},
		},
		{
			cmd: []string{"mkdir", ".vim/autoload/go"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vim/autoload/go/complete.vim --> b
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"rm", "-r", ".vim"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vim/autoload/ --> b
				link .vim/syntax/go.vim --> b
				link .vim/syntax/vim.vim --> a
				3 updated, 0 removed, 0 failed
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestUpdateError(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"cd", "w"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				nzn: no repository found in '.*' \(\.nzn not found\)! (re)
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
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				nzn: \.nzn[/\\]state.json: unexpected end of JSON input (re)
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.bashrc"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.gitconfig"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/a/.vim/syntax"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.vim/syntax/vim.vim"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.vimrc"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "a"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .bashrc --> a
				link .gitconfig --> a
				link .vim/ --> a
				link .vimrc --> a
				4 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/.gitconfig"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"rm", ".gitconfig"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .gitconfig --> b
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"touch", ".nzn/r/b/.vimrc"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"rm", ".vimrc"},
		},
		{
			cmd: []string{"touch", ".vimrc"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				nzn: .vimrc: not tracked
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".vimrc"},
		},
		{
			cmd: []string{"ln", "-s", ".nzn/r/b/.vimrc", ".vimrc"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink .vimrc -/- a
				nzn: not linked to layer 'a'
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".vimrc"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vimrc --> b
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"rm", ".vimrc"},
		},
		{
			cmd: []string{"touch", ".vimrc"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vimrc --> b
				error: .vimrc: .* (re)
				0 updated, 0 removed, 1 failed
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".vimrc"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/b/.vim/syntax"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/.vim/syntax/go.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink .vim/ -/- a
				link .vim/syntax/go.vim --> b
				link .vim/syntax/vim.vim --> a
				link .vimrc --> b
				3 updated, 1 removed, 0 failed
			`),
		},
		{
			cmd: []string{"rm", "-r", ".vim/syntax"},
		},
		{
			cmd: []string{"touch", ".vim/syntax"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vim/syntax/go.vim --> b
				error: .vim/syntax/go.vim: .* (re)
				link .vim/syntax/vim.vim --> a
				error: .vim/syntax/vim.vim: .* (re)
				0 updated, 0 removed, 2 failed
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".vim/syntax"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/_/.vim/syntax"},
		},
		{
			cmd: []string{"ln", "-s", ".nzn/r/_/.vim/syntax", ".vim/syntax"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vim/syntax/go.vim --> b
				error: .vim/syntax: (re)
				link .vim/syntax/vim.vim --> a
				error: .vim/syntax: (re)
				0 updated, 0 removed, 2 failed
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".vim/syntax"},
		},
		{
			cmd: []string{"rm", "-r", ".nzn/r/_"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/1"},
		},
		{
			cmd: []string{"touch", ".nzn/r/c/1/.screenrc"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/2"},
		},
		{
			cmd: []string{"touch", ".nzn/r/c/2/.tmux.conf"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "c"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				nzn: cannot resolve layer 'c':
				    1
				    2
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "c/1"},
		},
		{
			cmd: []string{"rm", "-r", ".nzn/r/c/1"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				nzn: .nzn/r/c/1/.screenrc: .* (re)
				[1]
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}
