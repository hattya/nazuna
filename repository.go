//
// nazuna :: repository.go
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

package nazuna

import (
	"fmt"
	"path/filepath"
)

type Repository struct {
	Layers []*Layer

	ui      UI
	vcs     *VCS
	nzndir  string
	repodir string
}

func OpenRepository(ui UI, path string) (*Repository, error) {
	rootdir, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	for !isDir(filepath.Join(rootdir, ".nzn")) {
		p := rootdir
		rootdir = filepath.Dir(rootdir)
		if rootdir == p {
			return nil, fmt.Errorf("no repository found in '%s' (.nzn not found)!", path)
		}
	}

	nzndir := filepath.Join(rootdir, ".nzn")
	repodir := filepath.Join(nzndir, "repo")
	vcs, err := VCSFor(repodir)
	if err != nil {
		return nil, err
	}
	r := &Repository{
		ui:      ui,
		vcs:     vcs,
		nzndir:  nzndir,
		repodir: repodir,
	}

	if unmarshal(filepath.Join(r.repodir, "nazuna.json"), &r.Layers) != nil {
		r.Layers = []*Layer{}
	}
	return r, nil
}

func (r *Repository) Flush() error {
	return marshal(filepath.Join(r.repodir, "nazuna.json"), r.Layers)
}

func (r *Repository) NewLayer(name string) (*Layer, error) {
	found := false
	for _, l := range r.Layers {
		if l.Name == name {
			found = true
		}
	}
	if found || !isEmptyDir(filepath.Join(r.repodir, name)) {
		return nil, fmt.Errorf("layer '%s' already exists!", name)
	}

	l := &Layer{
		Name: name,
	}
	r.Layers = append(r.Layers, nil)
	copy(r.Layers[1:], r.Layers)
	r.Layers[0] = l
	return l, r.Flush()
}

func (r *Repository) Add(paths ...string) error {
	cmd := r.vcs.Add(paths...)
	cmd.Dir = r.repodir
	return r.ui.Exec(cmd)
}
