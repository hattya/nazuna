//
// nazuna :: repository.go
//
//   Copyright (c) 2013-2014 Akinori Hattori <hattya@gmail.com>
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

package nazuna

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// disable repository discovery in tests
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
	r := &Repository{
		ui:      ui,
		vcs:     vcs,
		nzndir:  nzndir,
		rdir:    rdir,
		subroot: filepath.Join(nzndir, "sub"),
	}

	path = filepath.Join(r.rdir, "nazuna.json")
	if _, err := os.Stat(path); err == nil {
		if err := unmarshal(path, &r.Layers); err != nil {
			return nil, err
		}
	} else {
		r.Layers = []*Layer{}
	}
	return r, nil
}

func (r *Repository) Flush() error {
	return marshal(filepath.Join(r.rdir, "nazuna.json"), r.Layers)
}

func (r *Repository) LayerOf(name string) (*Layer, error) {
	n, err := r.splitLayer(name)
	if err != nil {
		return nil, err
	}
	for _, l := range r.Layers {
		if n[0] == l.Name {
			switch {
			case len(n) == 1:
				return l, nil
			case len(l.Layers) == 0:
				return nil, fmt.Errorf("layer '%v' is not abstract", n[0])
			}
			for _, ll := range l.Layers {
				if n[1] == ll.Name {
					ll.abstract = l
					return ll, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("layer '%v' does not exist!", name)
}

func (r *Repository) NewLayer(name string) (*Layer, error) {
	switch _, err := r.LayerOf(name); {
	case err != nil && !strings.Contains(err.Error(), "not exist"):
		return nil, err
	case err == nil || !IsEmptyDir(filepath.Join(r.rdir, name)):
		return nil, fmt.Errorf("layer '%v' already exists!", name)
	}

	var l *Layer
	newLayer := func(n string) *Layer {
		r.Layers = append(r.Layers, nil)
		copy(r.Layers[1:], r.Layers)
		r.Layers[0] = &Layer{Name: n}
		return r.Layers[0]
	}
	switch n, _ := r.splitLayer(name); len(n) {
	case 1:
		l = newLayer(n[0])
	default:
		var err error
		l, err = r.LayerOf(n[0])
		if err != nil {
			l = newLayer(n[0])
		}
		ll := &Layer{
			Name:     n[1],
			abstract: l,
		}
		l.Layers = append(l.Layers, ll)
		sort.Sort(layerByName(l.Layers))
		l = ll
	}
	os.MkdirAll(r.PathFor(l, "/"), 0777)
	return l, nil
}

func (r *Repository) splitLayer(name string) ([]string, error) {
	n := strings.Split(name, "/")
	if 2 < len(n) || strings.TrimSpace(n[0]) == "" || (1 < len(n) && strings.TrimSpace(n[1]) == "") {
		return nil, fmt.Errorf("invalid layer '%v'", name)
	}
	return n, nil
}

func (r *Repository) PathFor(layer *Layer, path string) string {
	if layer != nil {
		return filepath.Join(r.rdir, layer.Path(), path)
	}
	return filepath.Join(r.rdir, path)
}

func (r *Repository) SubrepoFor(path string) string {
	return filepath.Join(r.subroot, path)
}

func (r *Repository) WC() (*WC, error) {
	return openWC(r.ui, r)
}

func (r *Repository) Find(layer *Layer, path string) (typ string) {
	err := r.Walk(r.PathFor(layer, path), func(p string, fi os.FileInfo, err error) error {
		if !strings.HasSuffix(p, "/"+path) {
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

func (r *Repository) Walk(path string, walk filepath.WalkFunc) error {
	cmd := r.vcs.List(path)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()
	defer cmd.Process.Kill()
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		l := scanner.Text()
		fi, err := os.Stat(filepath.Join(r.rdir, l))
		if err = walk(l, fi, err); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (r *Repository) Add(paths ...string) error {
	return r.vcs.Add(paths...)
}

func (r *Repository) Command(args ...string) error {
	return r.vcs.Exec(args...)
}
