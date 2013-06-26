//
// nazuna :: subrepo_test.go
//
//   Copyright (c) 2013 Akinori Hattori <hattya@gmail.com>
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

package nazuna_test

import (
	"testing"
)

func TestSubrepo(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"git", "init", "-q", "r"},
		},
		{
			cmd: []string{"cd", "r"},
		},
		{
			cmd: []string{"touch", "_"},
		},
		{
			cmd: []string{"git", "add", "."},
		},
		{
			cmd: []string{"git", "-c", "user.email=nazuna@example.com", "commit", "-qm."},
		},
		{
			cmd: []string{"cd", ".."},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git", "w"},
		},
		{
			cmd: []string{"cd", "w"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "_", "r"},
		},
		{
			cmd: []string{"mkdir", ".nzn/sub"},
		},
		{
			cmd: []string{"git", "clone", "-q", "../r", ".nzn/sub/_"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `link r --> _
1 updated, 0 removed, 0 failed
`,
		},
		{
			cmd: []string{"ls", "r"},
			out: `.git/
_
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}

func TestSubrepoError(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"nzn", "subrepo"},
			out: `nzn: no repository found in '.*' \(\.nzn not found\)! (re)
[1]
`,
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"touch", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "subrepo"},
			out: `nzn: unexpected end of JSON input
[1]
`,
		},
		{
			cmd: []string{"rm", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "subrepo"},
			out: `nzn subrepo: flag --layer is required
usage: nzn subrepo -l <layer> -a <repository> <path>

manage subrepositories

  subrepo is used to manage external repositories.

  subrepo can associate <repository> to <path> by flag --add. If <path> ends
  with a path separator, it will be associated as the basename of <repository>
  under <path>.

options:

  -l, --layer    a layer
  -a, --add      add <repository> to <path>

[2]
`,
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a"},
			out: `nzn: invalid arguments
[1]
`,
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a"},
			out: `nzn: invalid arguments
[1]
`,
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
			out: `nzn: layer 'a' does not exist!
[1]
`,
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/a/.vim/bundle/vim-pathogen/autoload"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.vim/bundle/vim-pathogen/autoload/pathogen.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "a"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
			out: `nzn: '.vim/bundle/vim-pathogen' already exists!
[1]
`,
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-rfq", "a/.vim"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", "../dst"},
			out: `nzn: '../dst' is not under root
[1]
`,
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
			out: `nzn: subrepo '.vim/bundle/vim-pathogen' already exists!
[1]
`,
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/kien/ctrlp.vim", ".vim/bundle/ctrlp.vim"},
		},
		{
			cmd: []string{"mkdir", ".nzn/sub/github.com/kien/ctrlp.vim"},
		},
		{
			cmd: []string{"mkdir", ".nzn/sub/github.com/tpope/vim-pathogen"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `link .vim/bundle/ctrlp.vim --> github.com/kien/ctrlp.vim
link .vim/bundle/vim-pathogen --> github.com/tpope/vim-pathogen
2 updated, 0 removed, 0 failed
`,
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "b", "-a", "github.com/kien/ctrlp.vim", ".vim/bundle/"},
		},
		{
			cmd: []string{"rm", ".vim/bundle/ctrlp.vim"},
		},
		{
			cmd: []string{"mkdir", ".nzn/sub/_"},
		},
		{
			cmd: []string{"ln", "-s", ".nzn/sub/_", ".vim/bundle/ctrlp.vim"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `unlink .vim/bundle/ctrlp.vim -/- github.com/kien/ctrlp.vim
nzn: not linked to 'github.com/kien/ctrlp.vim'
[1]
`,
		},
		{
			cmd: []string{"rm", ".vim/bundle/ctrlp.vim"},
		},
		{
			cmd: []string{"rm", ".nzn/sub/_"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/b/.vim/bundle/ctrlp.vim/plugin"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/.vim/bundle/ctrlp.vim/plugin/ctrlp.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `warning: subrepo: '.vim/bundle/ctrlp.vim' exists in the repository
link .vim/bundle/ctrlp.vim/ --> b
1 updated, 0 removed, 0 failed
`,
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/1"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "c", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
			out: `nzn: layer 'c' is abstract
[1]
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}
