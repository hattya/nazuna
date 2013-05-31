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
	"os"
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func TestVCS(t *testing.T) {
	dir, err := mkdtemp()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	rc, _, berr := runCLI("nazuna.test", "init", "--vcs=git", dir)
	if rc != 0 {
		t.Logf("stderr:\n%s", berr)
		t.Fatalf("expected 0, got %d", rc)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rc, bout, berr := runCLI("nazuna.test", "vcs", "--version")
	if rc != 0 {
		t.Errorf("expected 0, got %d", rc)
	}
	if !strings.HasPrefix(bout, "git version ") {
		t.Errorf(`expected "git version .*", got %q`, bout)
	}
	if berr != "" {
		t.Errorf(`expected "", got %q`, berr)
	}
}

func TestVCSError(t *testing.T) {
	dir, err := mkdtemp()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rc, bout, berr := runCLI("nazuna.test", "vcs", "--version")
	if rc != 1 {
		t.Errorf("expected 1, got %d", rc)
	}
	if bout != "" {
		t.Errorf(`expected "", got %q`, bout)
	}
	if !strings.Contains(berr, ": no repository found ") {
		t.Errorf("error expected")
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
	defer os.RemoveAll(dir)

	_, err = nazuna.VCSFor(dir)
	if err == nil {
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
