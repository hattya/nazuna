//
// nazuna :: vcs_test.go
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

package nazuna_test

import (
	"os"
	"testing"

	"github.com/hattya/nazuna"
)

type testVCS struct {
	nazuna.BaseVCS
}

func newTest(ui nazuna.UI, dir string) nazuna.VCS {
	return &testVCS{nazuna.BaseVCS{
		Name: "name",
		Cmd:  "cmd",
		UI:   ui,
		Dir:  dir,
	}}
}

func init() {
	nazuna.RegisterVCS("test", ".test", newTest)
}

func TestRegisterVCSPanic(t *testing.T) {
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected panic")
			}
		}()

		nazuna.RegisterVCS("test", ".test", newTest)
	}()

	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected panic")
			}
		}()

		nazuna.RegisterVCS("TEST", ".test", newTest)
	}()
}

func TestFindVCS(t *testing.T) {
	vcs, err := nazuna.FindVCS(nil, "hg", "")
	if err != nil {
		t.Fatal(err)
	}
	switch vcs := vcs.(type) {
	case *nazuna.Mercurial:
		if g, e := vcs.Dir, ""; g != e {
			t.Errorf("VCS.Dir = %q, expected %q", g, e)
		}
	default:
		t.Errorf("expected *Mercurial, got %T", vcs)
	}

	if _, err := nazuna.FindVCS(nil, "cvs", ""); err == nil {
		t.Error("expected error")
	}
}

func TestVCSFor(t *testing.T) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	popd, err := pushd(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	if err := mkdir(".git"); err != nil {
		t.Fatal(err)
	}
	vcs, err := nazuna.VCSFor(nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	switch vcs := vcs.(type) {
	case *nazuna.Git:
		if g, e := vcs.Dir, dir; g != e {
			t.Errorf("VCS.Dir = %q, expected %q", g, e)
		}
	default:
		t.Fatalf("expected *Git, got %T", vcs)
	}

	if err := os.Remove(".git"); err != nil {
		t.Fatal(err)
	}
	if _, err = nazuna.VCSFor(nil, dir); err == nil {
		t.Error("expected error")
	}
}

func TestBaseVCS(t *testing.T) {
	vcs, err := nazuna.FindVCS(nil, "test", "")
	if err != nil {
		t.Fatal(err)
	}

	if g, e := vcs.String(), "name"; g != e {
		t.Errorf("VCS.String() = %q, expected %q", g, e)
	}
	if err := vcs.Init("dir"); err == nil {
		t.Error("expected error")
	}
	if err := vcs.Clone("src", "dst"); err == nil {
		t.Error("expected error")
	}
	if err := vcs.Add("a", "b", "c"); err == nil {
		t.Error("expected error")
	}
	if cmd := vcs.List("a", "b", "c"); cmd != nil {
		t.Errorf("expected nil, got %T", cmd)
	}
	if err := vcs.Update(); err == nil {
		t.Error("expected error")
	}
}

func TestGitVCS(t *testing.T) {
	testVCSImpl(t, "git")
}

func TestMercurialVCS(t *testing.T) {
	testVCSImpl(t, "hg")
}

func testVCSImpl(t *testing.T, cmd string) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	popd, err := pushd(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	ui := new(testUI)
	vcs, err := nazuna.FindVCS(ui, cmd, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := vcs.Init("r"); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Clone("r", "w"); err != nil {
		t.Fatal(err)
	}

	vcs, err = nazuna.FindVCS(ui, cmd, "r")
	if err != nil {
		t.Fatal(err)
	}
	if err := vcs.Update(); err == nil {
		t.Error("expected error")
	}
	if err := touch("r", "file"); err != nil {
		t.Fatal(err)
	}
	if err := mkdir("r", "dir"); err != nil {
		t.Fatal(err)
	}
	if err := touch("r", "dir", "file"); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Add("."); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Exec("commit", "-m", "."); err != nil {
		t.Fatal(err)
	}

	vcs, err = nazuna.FindVCS(ui, cmd, "w")
	if err != nil {
		t.Fatal(err)
	}
	ui.Reset()
	if err := ui.Exec(vcs.List(".")); err != nil {
		t.Fatal(err)
	}
	if g, e := ui.String(), ""; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if err := vcs.Update(); err != nil {
		t.Fatal(err)
	}
	ui.Reset()
	if err := ui.Exec(vcs.List(".")); err != nil {
		t.Fatal(err)
	}
	if g, e := ui.String(), "dir/file\nfile\n"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}
