//
// nazuna :: repository.go
//
//   Copyright (c) 2013-2021 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var discover = true

func Discover(b bool) bool {
	old := discover
	discover = b
	return old
}

type Repository struct {
	Layers []*Layer

	ui      UI
	vcs     VCS
	root    string
	nzndir  string
	rdir    string
	subroot string
}

func Open(ui UI, path string) (*Repository, error) {
	root, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	for !IsDir(filepath.Join(root, ".nzn")) {
		p := root
		root = filepath.Dir(root)
		if !discover || root == p {
			return nil, fmt.Errorf("no repository found in '%v' (.nzn not found)!", path)
		}
	}

	nzndir := filepath.Join(root, ".nzn")
	rdir := filepath.Join(nzndir, "r")
	vcs, err := VCSFor(ui, rdir)
	if err != nil {
		return nil, err
	}
	repo := &Repository{
		ui:      ui,
		vcs:     vcs,
		root:    root,
		nzndir:  nzndir,
		rdir:    rdir,
		subroot: filepath.Join(nzndir, "sub"),
	}

	if err := unmarshal(repo, filepath.Join(repo.rdir, "nazuna.json"), &repo.Layers); err != nil {
		return nil, err
	}
	if repo.Layers == nil {
		repo.Layers = []*Layer{}
	}
	return repo, nil
}

func (repo *Repository) Flush() error {
	return marshal(repo, filepath.Join(repo.rdir, "nazuna.json"), repo.Layers)
}

func (repo *Repository) LayerOf(name string) (*Layer, error) {
	n, err := repo.splitLayer(name)
	if err != nil {
		return nil, err
	}
	for _, l := range repo.Layers {
		if n[0] == l.Name {
			switch {
			case len(n) == 1:
				l.repo = repo
				return l, nil
			case len(l.Layers) == 0:
				return nil, fmt.Errorf("layer '%v' is not abstract", n[0])
			}
			for _, ll := range l.Layers {
				if n[1] == ll.Name {
					ll.repo = repo
					ll.abst = l
					return ll, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("layer '%v' does not exist!", name)
}

func (repo *Repository) NewLayer(name string) (*Layer, error) {
	switch _, err := repo.LayerOf(name); {
	case err != nil && !strings.Contains(err.Error(), "not exist"):
		return nil, err
	case err == nil || !IsEmptyDir(filepath.Join(repo.rdir, name)):
		return nil, fmt.Errorf("layer '%v' already exists!", name)
	}

	var l *Layer
	switch n, _ := repo.splitLayer(name); len(n) {
	case 1:
		l = repo.newLayer(n[0])
	default:
		var err error
		l, err = repo.LayerOf(n[0])
		if err != nil {
			l = repo.newLayer(n[0])
		}
		ll := &Layer{
			Name: n[1],
			repo: repo,
			abst: l,
		}
		l.Layers = append(l.Layers, ll)
		sort.Slice(l.Layers, func(i, j int) bool { return l.Layers[i].Name < l.Layers[j].Name })
		l = ll
	}
	os.MkdirAll(repo.PathFor(l, "/"), 0o777)
	return l, nil
}

func (repo *Repository) newLayer(name string) *Layer {
	repo.Layers = append(repo.Layers, nil)
	copy(repo.Layers[1:], repo.Layers)
	repo.Layers[0] = &Layer{
		Name: name,
		repo: repo,
	}
	return repo.Layers[0]
}

func (repo *Repository) splitLayer(name string) ([]string, error) {
	n := strings.Split(name, "/")
	for i := range n {
		n[i] = strings.TrimSpace(n[i])
	}
	if n[0] == "" || (len(n) > 1 && n[1] == "") || len(n) > 2 {
		return nil, fmt.Errorf("invalid layer '%v'", name)
	}
	return n, nil
}

func (repo *Repository) PathFor(layer *Layer, path string) string {
	if layer == nil {
		return filepath.Join(repo.rdir, path)
	}
	return filepath.Join(repo.rdir, layer.Path(), path)
}

func (repo *Repository) SubrepoFor(path string) string {
	return filepath.Join(repo.subroot, path)
}

func (repo *Repository) WC() (*WC, error) {
	return openWC(repo)
}

func (repo *Repository) Find(layer *Layer, path string) (typ string) {
	err := repo.Walk(repo.PathFor(layer, path), func(p string, fi os.FileInfo, err error) error {
		if !strings.HasSuffix(p, "/"+filepath.ToSlash(path)) {
			typ = "dir"
		} else {
			typ = "file"
		}
		return filepath.SkipDir
	})
	if err == filepath.SkipDir {
		return
	}

	for _, dst := range layer.Aliases {
		if dst == path {
			return "alias"
		}
	}

	dir, name := SplitPath(path)
	for _, l := range layer.Links[dir] {
		if l.Dst == name {
			return "link"
		}
	}

	for _, s := range layer.Subrepos[dir] {
		if s.Name == name || filepath.Base(s.Src) == name {
			return "subrepo"
		}
	}
	return
}

func (repo *Repository) Walk(path string, walk filepath.WalkFunc) error {
	cmd := repo.vcs.List(path)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()
	defer cmd.Process.Kill()
	s := bufio.NewScanner(out)
	for s.Scan() {
		p := s.Text()
		fi, err := os.Stat(filepath.Join(repo.rdir, p))
		if err = walk(p, fi, err); err != nil {
			return err
		}
	}
	return s.Err()
}

func (repo *Repository) Add(paths ...string) error {
	return repo.vcs.Add(paths...)
}

func (repo *Repository) Command(args ...string) error {
	return repo.vcs.Exec(args...)
}
