//
// nazuna :: init_test.go
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

func TestInit(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs=git", "$tempdir"},
		},
		{
			cmd: []string{"ls", "$tempdir/.nzn"},
			out: `repo/
`,
		},
		{
			cmd: []string{"ls", "$tempdir/.nzn/repo"},
			out: `.git/
nazuna.json
`,
		},
		{
			cmd: []string{"cat", "$tempdir/.nzn/repo/nazuna.json"},
			out: `[]
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}

func TestInitError(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs=cvs", "$tempdir"},
			out: `nzn: unknown vcs 'cvs'
[1]
`,
		},
		{
			cmd: []string{"nzn", "init", "$tempdir"},
			out: `nzn init: flag --vcs is required
usage: nzn init --vcs=<type> [<path>]

  create a new repository in the specified directory

options:

      --vcs=<type>    vcs type

[2]
`,
		},
		{
			cmd: []string{"mkdir", "$tempdir/.nzn/repo"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs=git", "$tempdir"},
			out: `nzn: repository '.*' already exists! (re)
[1]
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}
