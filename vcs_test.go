//
// nazuna :: vcs_test.go
//
//   Copyright (c) 2013-2017 Akinori Hattori <hattya@gmail.com>
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
			switch err := recover(); {
			case err == nil:
				t.Fatal("does not panic")
			case err != "NewVCS is nil":
				t.Fatal("unexpected panic:", err)
			}
		}()
		nazuna.RegisterVCS("", "", nil)
	}()

	func() {
		defer func() {
			switch err := recover(); {
			case err == nil:
				t.Fatal("does not panic")
			case err != "vcs 'test' already registered":
				t.Fatal("unexpected panic:", err)
			}
		}()
		nazuna.RegisterVCS("test", ".test2", newTest)
	}()
}

func TestFindVCS(t *testing.T) {
	switch _, err := nazuna.FindVCS(nil, "cvs", ""); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "unknown vcs 'cvs'"):
		t.Error("unexpected error:", err)
	}

	vcs, err := nazuna.FindVCS(nil, "hg", "")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := vcs.String(), "Mercurial"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	hg, ok := vcs.(*nazuna.Mercurial)
	if !ok {
		t.Fatalf("expected *nazuna.Mercurial, got %T", vcs)
	}
	if g, e := hg.Cmd, "hg"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := hg.Dir, ""; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	vcs, err = nazuna.FindVCS(nil, "test", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := vcs.(*testVCS); !ok {
		t.Fatalf("expected *testVCS, got %T", vcs)
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

	if _, err = nazuna.VCSFor(nil, dir); err == nil {
		t.Error("expected error")
	}

	if err := mkdir(".git"); err != nil {
		t.Fatal(err)
	}
	vcs, err := nazuna.VCSFor(nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	if g, e := vcs.String(), "Git"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	git, ok := vcs.(*nazuna.Git)
	if !ok {
		t.Fatalf("expected *nazuna.Git, got %T", vcs)
	}
	if g, e := git.Cmd, "git"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	if g, e := git.Dir, dir; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	if err := os.Rename(".git", ".test"); err != nil {
		t.Fatal(err)
	}
	vcs, err = nazuna.VCSFor(nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := vcs.(*testVCS); !ok {
		t.Fatalf("expected *testVCS, got %T", vcs)
	}
}

func TestBaseVCS(t *testing.T) {
	vcs, err := nazuna.FindVCS(nil, "test", "")
	if err != nil {
		t.Fatal(err)
	}

	switch err := vcs.Init("dir"); {
	case err == nil:
		t.Error("expected error")
	case err.Error() != "VCS.Init not implemented":
		t.Error("unexpected error:", err)
	}
	switch err := vcs.Clone("src", "dst"); {
	case err == nil:
		t.Error("expected error")
	case err.Error() != "VCS.Clone not implemented":
		t.Error("unexpected error:", err)
	}
	switch err := vcs.Add("paths"); {
	case err == nil:
		t.Error("expected error")
	case err.Error() != "VCS.Add not implemented":
		t.Error("unexpected error:", err)
	}
	if c := vcs.List("paths"); c != nil {
		t.Errorf("expected nil, got %T", c)
	}
	switch err := vcs.Update(); {
	case err == nil:
		t.Error("expected error")
	case err.Error() != "VCS.Update not implemented":
		t.Error("unexpected error:", err)
	}
}

func TestGitVCS(t *testing.T) {
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

	ui := &testUI{}
	vcs, err := nazuna.FindVCS(ui, "git", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := vcs.Init("r"); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Clone("r", "w"); err != nil {
		t.Fatal(err)
	}

	vcs, err = nazuna.FindVCS(ui, "git", "r")
	if err != nil {
		t.Fatal(err)
	}
	if err := touch("r/file"); err != nil {
		t.Fatal(err)
	}
	if err := mkdir("r/dir"); err != nil {
		t.Fatal(err)
	}
	if err := touch("r/dir/file"); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Add("."); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Exec("commit", "-am", "."); err != nil {
		t.Fatal(err)
	}

	vcs, err = nazuna.FindVCS(ui, "git", "w")
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

func TestMercurialVCS(t *testing.T) {
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

	ui := &testUI{}
	vcs, err := nazuna.FindVCS(ui, "hg", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := vcs.Init("r"); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Clone("r", "w"); err != nil {
		t.Fatal(err)
	}

	vcs, err = nazuna.FindVCS(ui, "hg", "r")
	if err != nil {
		t.Fatal(err)
	}
	if err := touch("r/file"); err != nil {
		t.Fatal(err)
	}
	if err := mkdir("r/dir"); err != nil {
		t.Fatal(err)
	}
	if err := touch("r/dir/file"); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Add("."); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Exec("commit", "-m", "."); err != nil {
		t.Fatal(err)
	}

	vcs, err = nazuna.FindVCS(ui, "hg", "w")
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
