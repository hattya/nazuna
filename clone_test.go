//
// nazuna :: clone_test.go
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

func TestClone(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"git", "init", "-q", "src"},
		},
		{
			cmd: []string{"nzn", "clone", "--vcs=git", "src", "dst"},
			out: `Cloning into 'dst/.nzn/repo'...
warning: .* (re)
done.
`,
		},
		{
			cmd: []string{"ls", "dst/.nzn"},
			out: `repo/
`,
		},
		{
			cmd: []string{"ls", "dst/.nzn/repo"},
			out: `.git/
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}

func TestCloneError(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"nzn", "clone"},
			out: `nzn: invalid arguments
[1]
`,
		},
		{
			cmd: []string{"nzn", "clone", "src"},
			out: `nzn clone: flag -*vcs is required (re)
usage: nzn clone --vcs=<type> <repository> [<path>]

  make a copy of an existing repository

options:

      --vcs=<type>    vcs type

[2]
`,
		},
		{
			cmd: []string{"nzn", "clone", "--vcs=cvs", "src"},
			out: `nzn: unknown vcs 'cvs'
[1]
`,
		},
		{
			cmd: []string{"git", "init", "-q", "src"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs=git", "dst"},
		},
		{
			cmd: []string{"nzn", "clone", "--vcs=git", "src", "dst"},
			out: `nzn: repository 'dst' already exists!
[1]
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}
