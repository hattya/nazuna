//
// nazuna :: alias_test.go
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
	"path/filepath"
	"testing"
)

func TestAlias(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs=git"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"mkdir", ".nzn/repo/a/.config/gocode"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/a/.config/gocode/config.json"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/a/.gitconfig"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/a/.vimrc"},
		},
		{
			cmd: []string{"mkdir", ".nzn/repo/a/.vim/syntax"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/a/.vim/syntax/vim.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "a"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b/1"},
		},
		{
			cmd: []string{"mkdir", ".nzn/repo/b/1/.vim/syntax"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/b/1/.vim/syntax/go.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"nzn", "layer", "b/1"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `link .config/ --> a
link .gitconfig --> a
link .vim/syntax/go.vim --> b/1
link .vim/syntax/vim.vim --> a
link .vimrc --> a
5 updated, 0 removed, 0 failed
`,
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b/2"},
		},
		{
			cmd: []string{"mkdir", ".nzn/repo/b/2/vimfiles/syntax"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/b/2/vimfiles/syntax/go.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/2", ".vim", "vimfiles"},
		},
		{
			cmd: []string{"export", "APPDATA=$tempdir/AppData/Roaming"},
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
			out: `unlink .config/ -/- a
unlink .vim/syntax/go.vim -/- b/1
unlink .vim/syntax/vim.vim -/- a
link AppData/Roaming/gocode/ --> a:.config/gocode/
link vimfiles/syntax/go.vim --> b/2
link vimfiles/syntax/vim.vim --> a:.vim/syntax/vim.vim
3 updated, 3 removed, 0 failed
`,
		},
		{
			cmd: []string{"rm", "-r", "AppData/Roaming"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `link AppData/Roaming/ --> a:.config/
1 updated, 0 removed, 0 failed
`,
		},
		{
			cmd: []string{"touch", ".nzn/repo/a/.curlrc"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "a"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/2", ".curlrc", "$APPDATA/_curlrc"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `unlink AppData/Roaming/ -/- a:.config/
link AppData/Roaming/_curlrc --> a:.curlrc
link AppData/Roaming/gocode/ --> a:.config/gocode/
2 updated, 1 removed, 0 failed
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}

func TestAliasError(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"nzn", "alias"},
			out: `nzn: no repository found in '.*' \(\.nzn not found\)! (re)
[1]
`,
		},
		{
			cmd: []string{"nzn", "init", "--vcs=git"},
		},
		{
			cmd: []string{"touch", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "alias"},
			out: `nzn: unexpected end of JSON input
[1]
`,
		},
		{
			cmd: []string{"rm", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "alias"},
			out: `nzn alias: flag --layer is required
usage: nzn alias -l <layer> <src> <dst>

create an alias for the specified path

  Change the location of <src> to <dst>. <src> should be existed in the lower layer
  than <dst>, and <src> is treated as <dst> in the layer <layer>. If <src> does
  not match any locations on update, it will be ignored without error.

  You can refer environment variables in <dst>. Supported formats are ${var}
  and $var.

options:

  -l, --layer    a layer

[2]
`,
		},
		{
			cmd: []string{"nzn", "alias", "-l", "a", "src", "dst"},
			out: `nzn: layer 'a' does not exist!
[1]
`,
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "a"},
			out: `nzn: invalid arguments
[1]
`,
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b/1"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b", "src", "dst"},
			out: `nzn: layer 'b' is abstract
[1]
`,
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "../dst"},
			out: `nzn: '../dst' is not under root
[1]
`,
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "src"},
			out: `nzn: 'src' and 'src' are the same file
[1]
`,
		},
		{
			cmd: []string{"touch", ".nzn/repo/b/1/dst"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "dst"},
			out: `nzn: 'dst' already exists!
[1]
`,
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-fq", "b/1/dst"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "dst"},
		},
		{
			cmd: []string{"nzn", "alias", "-l", "b/1", "src", "dst"},
			out: `nzn: alias 'dst' already exists!
[1]
`,
		},
		{
			cmd: []string{"export", "ROOT=../"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/a/src"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "a"},
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
			out: `nzn: '` + filepath.Join("..", "dst") + `' is not under root
[1]
`,
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-fq", "a/src"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "_", "src"},
		},
		{
			cmd: []string{"touch", "_"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `warning: link: '` + filepath.Join("..", "dst") + `' is not under root
0 updated, 0 removed, 0 failed
`,
		},
		{
			cmd: []string{"rm", "_"},
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}
