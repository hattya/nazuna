//
// nazuna :: layer_test.go
//
//   Copyright (c) 2014-2018 Akinori Hattori <hattya@gmail.com>
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
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hattya/nazuna"
)

func TestLayer(t *testing.T) {
	l := &nazuna.Layer{Name: "layer"}
	if g, e := l.Path(), "layer"; g != e {
		t.Errorf("Layer.Path() = %v, expected %v", g, e)
	}

	l.Abst(&nazuna.Layer{Name: "abst"})
	if g, e := l.Path(), "abst/layer"; g != e {
		t.Errorf("Layer.Path() = %v, expected %v", g, e)
	}
}

func TestLayerNewAlias(t *testing.T) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	repo, err := create(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := repo.NewLayer("abst/layer"); err != nil {
		t.Fatal(err)
	}

	l, err := repo.LayerOf("abst")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.NewAlias("src", "dst"); err == nil {
		t.Error("expected error")
	}

	l, err = repo.LayerOf("abst/layer")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.NewAlias("src", "src"); err == nil {
		t.Error("expected error")
	}
	if err := l.NewAlias("src", "dst"); err != nil {
		t.Error(err)
	}

	if err := l.NewAlias("src", "dst"); err == nil {
		t.Error("expected error")
	}

	l.Aliases = nil
	if err := touch(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if err := l.NewAlias("src", "dst"); err == nil {
		t.Error("expected error")
	}

	l.Aliases = nil
	if err := repo.Command("rm", "-rf", "."); err != nil {
		t.Fatal(err)
	}
	if err := mkdir(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, filepath.Join("dst", "file"))); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if err := l.NewAlias("src", "dst"); err != nil {
		t.Error(err)
	}
}

func TestLayerNewLink(t *testing.T) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	repo, err := create(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := repo.NewLayer("abst/layer"); err != nil {
		t.Fatal(err)
	}

	l, err := repo.LayerOf("abst")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewLink([]string{}, "src", "dst"); err == nil {
		t.Error("expected error")
	}

	l, err = repo.LayerOf("abst/layer")
	if err != nil {
		t.Fatal(err)
	}
	lnk, err := l.NewLink([]string{"path"}, "src", "dst")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := lnk.Src, "src"; g != e {
		t.Errorf("Link.Src = %v, expected %v", g, e)
	}
	if g, e := lnk.Dst, "dst"; g != e {
		t.Errorf("Link.Dst = %v, expected %v", g, e)
	}

	if _, err := l.NewLink([]string{"path"}, "src", "dst"); err == nil {
		t.Error("expected error")
	}

	l.Links = nil
	if err := touch(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewLink([]string{"path"}, "src", "dst"); err == nil {
		t.Error("expected error")
	}

	l.Links = nil
	if err := repo.Command("rm", "-rf", "."); err != nil {
		t.Fatal(err)
	}
	if err := mkdir(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, filepath.Join("dst", "file"))); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewLink([]string{"path"}, "src", "dst"); err == nil {
		t.Error("expected error")
	}
}

func TestLayerNewSubrepo(t *testing.T) {
	dir, err := tempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	repo, err := create(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := repo.NewLayer("abst/layer"); err != nil {
		t.Fatal(err)
	}
	src := "github.com/hattya/nazuna"

	l, err := repo.LayerOf("abst")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.NewSubrepo(src, "dst"); err == nil {
		t.Error("expected error")
	}

	l, err = repo.LayerOf("abst/layer")
	if err != nil {
		t.Fatal(err)
	}
	sub, err := l.NewSubrepo(src, "dst")
	if err != nil {
		t.Fatal(err)
	}
	if g, e := sub.Src, src; g != e {
		t.Errorf("Subrepo.Src = %v, expected %v", g, e)
	}
	if g, e := sub.Name, "dst"; g != e {
		t.Errorf("Subrepo.Name = %v, expected %v", g, e)
	}

	l.Subrepos = nil
	sub, err = l.NewSubrepo(src, filepath.Base(src))
	if err != nil {
		t.Fatal(err)
	}
	if g, e := sub.Src, src; g != e {
		t.Errorf("Subrepo.Src = %v, expected %v", g, e)
	}
	if g, e := sub.Name, ""; g != e {
		t.Errorf("Subrepo.Name = %v, expected %v", g, e)
	}

	if _, err := l.NewSubrepo(src, filepath.Base(src)); err == nil {
		t.Error("expected error")
	}

	l.Subrepos = nil
	if err := touch(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err = l.NewSubrepo(src, "dst"); err == nil {
		t.Error("expected error")
	}

	l.Subrepos = nil
	if err := repo.Command("rm", "-rf", "."); err != nil {
		t.Fatal(err)
	}
	if err := mkdir(repo.PathFor(l, "dst")); err != nil {
		t.Fatal(err)
	}
	if err := touch(repo.PathFor(l, filepath.Join("dst", "file"))); err != nil {
		t.Fatal(err)
	}
	if err := repo.Command("add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err = l.NewSubrepo(src, "dst"); err == nil {
		t.Error("expected error")
	}
}

func create(path string) (*nazuna.Repository, error) {
	rdir := filepath.Join(path, ".nzn", "r")
	if err := mkdir(rdir); err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = rdir
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return nazuna.Open(&testUI{}, path)
}

func TestSortLayers(t *testing.T) {
	layers := []*nazuna.Layer{
		{Name: "b"},
		{Name: "a"},
	}
	nazuna.SortLayers(layers)
	if g, e := layers[0].Name, "a"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := layers[1].Name, "b"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
}

func TestSortLinks(t *testing.T) {
	links := []*nazuna.Link{
		{Dst: "b"},
		{Dst: "a"},
	}
	nazuna.SortLinks(links)
	if g, e := links[0].Dst, "a"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := links[1].Dst, "b"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
}

func TestSortSubrepos(t *testing.T) {
	subrepos := []*nazuna.Subrepo{
		{Src: "b"},
		{Src: "a"},
	}
	nazuna.SortSubrepos(subrepos)
	if g, e := subrepos[0].Src, "a"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := subrepos[1].Src, "b"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
}
