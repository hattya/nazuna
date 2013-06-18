//
// nazuna :: link_test.go
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
	"os"
	"testing"
)

func TestLink(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	defer func() {
		os.Setenv("GOPATH", gopath)
	}()
	os.Setenv("GOPATH", "")

	sep := string(os.PathListSeparator)
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"mkdir", "root"},
		},
		{
			cmd: []string{"mkdir", "root/go/misc/vim"},
		},
		{
			cmd: []string{"mkdir", "root/gocode/src/github.com/nsf/gocode/vim"},
		},
		{
			cmd: []string{"mkdir", "work"},
		},
		{
			cmd: []string{"cd", "work"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs=git"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/a/.vimrc"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "a"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "$tempdir/root/go/misc/vim", ".vim/bundle/golang"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "-p", "$GOPATH" + sep + "$tempdir/root/gocode", "src/github.com/nsf/gocode/vim", ".vim/bundle/gocode"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `link .vim/bundle/gocode/ --> .*/root/gocode/src/github.com/nsf/gocode/vim/ (re)
link .vim/bundle/golang/ --> .*/root/go/misc/vim/ (re)
link .vimrc --> a
3 updated, 0 removed, 0 failed
`,
		},
		{
			cmd: []string{"rm", "-r", "../root/go"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `unlink .vim/bundle/golang/ -/- .*/root/go/misc/vim/ (re)
0 updated, 1 removed, 0 failed
`,
		},
		{
			cmd: []string{"rm", "-r", "../root"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `unlink .vim/bundle/gocode/ -/- .*/root/gocode/src/github.com/nsf/gocode/vim/ (re)
0 updated, 1 removed, 0 failed
`,
		},
		{
			cmd: []string{"ls", "."},
			out: `.nzn/
.vimrc
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}

func TestLinkError(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"nzn", "link"},
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
			cmd: []string{"nzn", "link"},
			out: `nzn: unexpected end of JSON input
[1]
`,
		},
		{
			cmd: []string{"rm", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "link"},
			out: `nzn link: flag -*layer is required (re)
usage: nzn link -l <layer> [-p <path>] <src> <dst>

create a link for the specified path

  link is used to create a link of <src> to <dst>, and will be managed by
  update. If <src> is not found on update, it will ignore without error.

  The value of flag --path is a list of directories like PATH or GOPATH
  environment variables, and it is used to search <src>.

  You can refer environment variables in <path> and <src>. Supported formats
  are ${var} and $var.

options:

  -l, --layer    a layer
  -p, --path     a list of directories to search <src>

[2]
`,
		},
		{
			cmd: []string{"nzn", "link", "-l", "a"},
			out: `nzn: invalid arguments
[1]
`,
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
			out: `nzn: layer 'a' does not exist!
[1]
`,
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "../dst"},
			out: `nzn: '../dst' is not under root
[1]
`,
		},
		{
			cmd: []string{"touch", ".nzn/repo/a/dst"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "a"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
			out: `nzn: 'dst' already exists!
[1]
`,
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-fq", "a/dst"},
		},
		{
			cmd: []string{"mkdir", ".nzn/repo/a/dst"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/a/dst/1"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "a"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
			out: `nzn: 'dst' already exists!
[1]
`,
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-frq", "a/dst"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "src", "dst"},
			out: `nzn: link 'dst' already exists!
[1]
`,
		},
		{
			cmd: []string{"touch", "src"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `link dst --> src
1 updated, 0 removed, 0 failed
`,
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
			out: `unlink dst -/- src
nzn: not linked to 'src'
[1]
`,
		},
		{
			cmd: []string{"rm", "dst"},
		},
		{
			cmd: []string{"rm", "_"},
		},
		{
			cmd: []string{"touch", ".nzn/repo/b/dst"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "b"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: `warning: link: 'dst' exists in the repository
link dst --> b
1 updated, 0 removed, 0 failed
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}
