//
// nazuna :: repository_test.go
//
//   Copyright (c) 2013-2022 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hattya/nazuna"
)

func TestOpen(t *testing.T) {
	sandbox(t)

	if err := mkdir(".nzn", "r", ".git"); err != nil {
		t.Fatal(err)
	}
	repo, err := nazuna.Open(nil, ".")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Flush(); err != nil {
		t.Error(err)
	}
	data, err := os.ReadFile(filepath.Join(".nzn", "r", "nazuna.json"))
	if err != nil {
		t.Fatal(err)
	}
	if g, e := string(data), "[]\n"; g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}

func TestOpenError(t *testing.T) {
	sandbox(t)

	// no repository
	switch _, err := nazuna.Open(nil, "."); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "no repository found "):
		t.Error("unexpected error:", err)
	}
	// unknown vcs
	if err := mkdir(".nzn", "r"); err != nil {
		t.Fatal(err)
	}
	switch _, err := nazuna.Open(nil, "."); {
	case err == nil:
		t.Error("expected error")
	case !strings.HasPrefix(err.Error(), "unknown vcs for directory "):
		t.Error("unexpected error:", err)
	}
	// unmarshal error
	if err := mkdir(".nzn", "r", ".git"); err != nil {
		t.Fatal(err)
	}
	if err := mkdir(".nzn", "r", "nazuna.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := nazuna.Open(nil, "."); err == nil {
		t.Error("expected error")
	}
}

func TestNewLayer(t *testing.T) {
	repo := init_(t)

	layers := []*nazuna.Layer{
		{Name: "a"},
		{Name: "z"},
	}
	for _, l := range layers {
		l.SetRepo(repo)
	}
	if _, err := repo.NewLayer("z"); err != nil {
		t.Error(err)
	}
	if _, err := repo.NewLayer("a"); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(repo.Layers, layers) {
		t.Error("unexpected order")
	}

	repo.Layers = nil
	layers = []*nazuna.Layer{{Name: "abst"}}
	abst := layers[0]
	abst.Layers = []*nazuna.Layer{
		{Name: "a"},
		{Name: "z"},
	}
	abst.SetRepo(repo)
	for _, l := range abst.Layers {
		l.SetRepo(repo)
		l.SetAbst(abst)
	}
	if _, err := repo.NewLayer("abst/z"); err != nil {
		t.Error(err)
	}
	if _, err := repo.NewLayer("abst/a"); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(repo.Layers, layers) {
		t.Error("expected to sort by Layer.Name")
	}
}

func TestNewLayerError(t *testing.T) {
	repo := init_(t)

	if _, err := repo.NewLayer("layer"); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{
		"",
		"/",
		"//",
		"layer/layer",
		"layer",
	} {
		if _, err := repo.NewLayer(n); err == nil {
			t.Error("expected error")
		}
	}
}

func TestRepositoryPaths(t *testing.T) {
	repo := init_(t)

	rdir := filepath.Join(repo.Root(), ".nzn", "r")
	if g, e := repo.PathFor(nil, "file"), filepath.Join(rdir, "file"); g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
	l, err := repo.NewLayer("layer")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := repo.PathFor(l, "file"), filepath.Join(rdir, l.Path(), "file"); g != e {
		t.Errorf("expected %q, got %q", e, g)
	}

	subroot := filepath.Join(repo.Root(), ".nzn", "sub")
	if g, e := repo.SubrepoFor("subrepo"), filepath.Join(subroot, "subrepo"); g != e {
		t.Errorf("expected %q, got %q", e, g)
	}
}

var findPathTests = []struct {
	typ, path string
}{
	{"", "_"},
	{"dir", "dir"},
	{"file", filepath.Join("dir", "file")},
	{"alias", "alias"},
	{"link", "link"},
	{"subrepo", "subrepo"},
}

func TestFindPath(t *testing.T) {
	repo := init_(t)

	l, err := repo.NewLayer("layer")
	if err != nil {
		t.Fatal(err)
	}
	if err := mkdir(repo.PathFor(l, "dir")); err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, filepath.Join("dir", "file"))); err != nil {
		t.Fatal(err)
	}
	if err := repo.Add("."); err != nil {
		t.Fatal(err)
	}
	if err := l.NewAlias("a", "alias"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewLink(nil, "l", "link"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewSubrepo("github.com/hattya/nazuna", "subrepo"); err != nil {
		t.Fatal(err)
	}

	for _, tt := range findPathTests {
		if g, e := repo.Find(l, tt.path), tt.typ; g != e {
			t.Errorf("expected %v, got %v", e, g)
		}
	}
}

func init_(t *testing.T) *nazuna.Repository {
	t.Helper()

	dir := sandbox(t)
	rdir := filepath.Join(dir, ".nzn", "r")
	if err := mkdir(rdir); err != nil {
		t.Fatal("init:", err)
	}
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = rdir
	if err := cmd.Run(); err != nil {
		t.Fatal("init:", err)
	}
	repo, err := nazuna.Open(new(testUI), dir)
	if err != nil {
		t.Fatal("init:", err)
	}
	return repo
}
