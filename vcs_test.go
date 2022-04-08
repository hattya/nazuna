//
// nazuna :: vcs_test.go
//
//   Copyright (c) 2013-2022 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna_test

import (
	"os"
	"path/filepath"
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
	dir := t.TempDir()
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
	testVCSImpl(t, "git", func(vcs nazuna.VCS) (err error) {
		for n, v := range map[string]string{
			"user.name":  "Nazuna",
			"user.email": "nazuna@example.com",
		} {
			if err = vcs.Exec("config", "--local", n, v); err != nil {
				break
			}
		}
		return
	})
}

func TestMercurialVCS(t *testing.T) {
	testVCSImpl(t, "hg", func(vcs nazuna.VCS) error {
		dir := vcs.(*nazuna.Mercurial).Dir
		data := "[ui]\nusername = Nazuna <nazuna@example.com>\n"
		return os.WriteFile(filepath.Join(dir, ".hg", "hgrc"), []byte(data), 0o666)
	})
}

func testVCSImpl(t *testing.T, cmd string, config func(nazuna.VCS) error) {
	popd, err := pushd(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer popd()

	ui := new(testUI)
	vcs, err := nazuna.FindVCS(ui, cmd, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := vcs.Init("repo"); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Clone("repo", "wc"); err != nil {
		t.Fatal(err)
	}

	vcs, err = nazuna.FindVCS(ui, cmd, "repo")
	if err != nil {
		t.Fatal(err)
	}
	if err := vcs.Update(); err == nil {
		t.Error("expected error")
	}
	if err := config(vcs); err != nil {
		t.Fatal(err)
	}
	if err := touch("repo", "file"); err != nil {
		t.Fatal(err)
	}
	if err := mkdir("repo", "dir"); err != nil {
		t.Fatal(err)
	}
	if err := touch("repo", "dir", "file"); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Add("."); err != nil {
		t.Fatal(err)
	}
	if err := vcs.Exec("commit", "-m", "."); err != nil {
		t.Fatal(err)
	}

	vcs, err = nazuna.FindVCS(ui, cmd, "wc")
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
