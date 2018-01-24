//
// nazuna/cmd/nzn :: clone_test.go
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
	"path/filepath"
	"testing"

	"github.com/hattya/go.cli"
)

func TestClone(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"git", "init", "-q", "r"},
		},
		{
			cmd: []string{"cd", "r"},
		},
		{
			cmd: []string{"touch", "nazuna.json"},
		},
		{
			cmd: []string{"git", "add", "."},
		},
		{
			cmd: []string{"git", "commit", "-qm."},
		},
		{
			cmd: []string{"cd", ".."},
		},
		{
			cmd: []string{"nzn", "clone", "--vcs", "git", "r", "w"},
			out: fmt.Sprintf(cli.Dedent(`
				Cloning into '%v'...
				done.
			`), filepath.FromSlash("w/.nzn/r")),
		},
		{
			cmd: []string{"ls", "w/.nzn"},
			out: cli.Dedent(`
				r/
			`),
		},
		{
			cmd: []string{"ls", "w/.nzn/r"},
			out: cli.Dedent(`
				.git/
				nazuna.json
			`),
		},
	}
	if err := s.exec(); err != nil {
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
			cmd: []string{"git", "init", "-q", "r"},
		},
		{
			cmd: []string{"nzn", "clone", "r", "w"},
			out: cli.Dedent(`
				nzn clone: --vcs flag is required (re)
				usage: nzn clone --vcs <type> <repository> [<path>]

				create a copy of an existing repository

				  Create a copy of an existing repository in <path>. If <path> does not exist,
				  it will be created.

				  If <path> is not specified, the current working diretory is used.

				options:

				  --vcs <type>    vcs type

				[2]
			`),
		},
		{
			cmd: []string{"nzn", "clone", "--vcs", "cvs", "r", "w"},
			out: cli.Dedent(`
				nzn: unknown vcs 'cvs'
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git", "w"},
		},
		{
			cmd: []string{"nzn", "clone", "--vcs", "git", "r", "w"},
			out: cli.Dedent(`
				nzn: repository 'w' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"cd", "w"},
		},
		{
			cmd: []string{"nzn", "clone", "--vcs", "git", "../r"},
			out: cli.Dedent(`
				nzn: repository '.' already exists!
				[1]
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}
