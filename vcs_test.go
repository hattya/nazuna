//
// nazuna :: vcs_test.go
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
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func TestVCS(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"nzn", "vcs", "--version"},
			out: `git version \d.* (re)
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}

func TestVCSError(t *testing.T) {
	ts := testScript{
		{
			cmd: []string{"mkdtemp"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"nzn", "vcs", "--version"},
			out: `nzn: no repository found in '.*' \(\.nzn not found\)! (re)
[1]
`,
		},
	}
	if err := ts.run(); err != nil {
		t.Error(err)
	}
}

func TestFindVCS(t *testing.T) {
	vcses := nazuna.VCSes
	defer func() {
		nazuna.VCSes = vcses
	}()

	nazuna.VCSes = []*nazuna.VCS{
		{
			Name: "Subversion",
			Cmd:  "svn",
		},
		{
			Name: "SVK",
			Cmd:  "svk",
		},
	}

	if _, err := nazuna.FindVCS("cvs"); err == nil || !strings.HasPrefix(err.Error(), "unknown vcs 'cvs'") {
		t.Error("error expected")
	}
	if _, err := nazuna.FindVCS("sv"); err == nil || !strings.HasPrefix(err.Error(), "vcs 'sv' is ambiguous:") {
		t.Error("error expected")
	}

	vcs, err := nazuna.FindVCS("svn")
	if err != nil {
		t.Fatal(err)
	}
	if vcs.Cmd != "svn" {
		t.Errorf(`expected "svn", got %q`, vcs.Cmd)
	}
	if vcs.String() != "Subversion" {
		t.Errorf(`expected "Subversion", got %q`, vcs)
	}
}

func TestVCSFor(t *testing.T) {
	dir, err := mkdtemp()
	if err != nil {
		t.Fatal(err)
	}
	defer nazuna.RemoveAll(dir)

	if _, err = nazuna.VCSFor(dir); err == nil {
		t.Error("error expected")
	}

	if err := mkdir(dir, ".git"); err != nil {
		t.Fatal(err)
	}
	vcs, err := nazuna.VCSFor(dir)
	if err != nil {
		t.Fatal(err)
	}
	if vcs.Cmd != "git" {
		t.Errorf(`expected "git", got %q`, vcs.Cmd)
	}
	if vcs.String() != "Git" {
		t.Errorf(`expected "Git", got %q`, vcs)
	}
}
