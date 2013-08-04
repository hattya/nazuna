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
			cmd: []string{"nzn", "vcs", "--version"},
			out: `git version \d.* (re)
`,
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestVCSError(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"cd", "w"},
		},
		{
			cmd: []string{"nzn", "vcs", "--version"},
			out: `nzn: no repository found in '.*' \(\.nzn not found\)! (re)
[1]
`,
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestFindVCS(t *testing.T) {
	v := nazuna.VCSes
	defer func() { nazuna.VCSes = v }()
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

	switch _, err := nazuna.FindVCS("cvs"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "unknown vcs 'cvs'"):
		t.Error("unexpected error:", err)
	}
	switch _, err := nazuna.FindVCS("sv"); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "vcs 'sv' is ambiguous:"):
		t.Error("unexpected error:", err)
	}

	vcs, err := nazuna.FindVCS("svn")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := vcs.Cmd, "svn"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := vcs.String(), "Subversion"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}

func TestVCSFor(t *testing.T) {
	dir, err := mkdtemp()
	if err != nil {
		t.Fatal(err)
	}
	defer nazuna.RemoveAll(dir)

	if _, err = nazuna.VCSFor(dir); err == nil {
		t.Error("expected error")
	}

	if err := mkdir(dir, ".git"); err != nil {
		t.Fatal(err)
	}
	vcs, err := nazuna.VCSFor(dir)
	if err != nil {
		t.Fatal(err)
	}
	if g, e := vcs.Cmd, "git"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := vcs.String(), "Git"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}
