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
	"bufio"
	"fmt"
	"io"
	"os"
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

	path = filepath.Join(r.repodir, "nazuna.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if err := unmarshal(path, &r.Layers); err != nil {
			return nil, err
		}
	} else {
		r.Layers = []*Layer{}
	}
	return r, nil
}

func (r *Repository) Flush() error {
	return marshal(filepath.Join(r.repodir, "nazuna.json"), r.Layers)
}

func (r *Repository) LayerOf(name string) (*Layer, error) {
	for _, l := range r.Layers {
		if name == l.Name {
			return l, nil
		}
	}
	return nil, fmt.Errorf("layer '%s' does not exist!", name)
}

func (r *Repository) NewLayer(name string) (*Layer, error) {
	if _, err := r.LayerOf(name); err == nil || !isEmptyDir(filepath.Join(r.repodir, name)) {
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

func (r *Repository) PathFor(layer *Layer, path string) string {
	return filepath.Join(r.repodir, layer.Name, path)
}

func (r *Repository) WC() (*WC, error) {
	return openWC(r.ui, r)
}

func (r *Repository) Walk(path string, walk filepath.WalkFunc) error {
	cmd := r.vcs.List(path)
	cmd.Dir = r.repodir
	pout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	out := bufio.NewReader(pout)
	var line []byte
	for {
		data, isPrefix, err := out.ReadLine()
		switch {
		case err == io.EOF:
			return cmd.Wait()
		case err != nil:
			cmd.Wait()
			return err
		default:
			line = append(line, data...)
			if isPrefix {
				continue
			}
		}
		p := string(line)
		fi, err := os.Stat(filepath.Join(r.repodir, p))
		if err := walk(p[len(path)+1:], fi, err); err != nil {
			return err
		}
		line = line[:0]
	}
}

func (r *Repository) Add(paths ...string) error {
	cmd := r.vcs.Add(paths...)
	cmd.Dir = r.repodir
	return r.ui.Exec(cmd)
}

func (r *Repository) Command(args ...string) error {
	cmd := r.vcs.Command(args...)
	cmd.Dir = r.repodir
	return r.ui.Exec(cmd)
}
