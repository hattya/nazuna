//
// nazuna :: wc.go
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
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

type ResolveError struct {
	Name string
	List []string
}

func (e *ResolveError) Error() string {
	s := fmt.Sprintf("cannot resolve layer '%s'", e.Name)
	if len(e.List) == 0 {
		return s
	}
	return fmt.Sprintf("%s:\n    %s", s, strings.Join(e.List, "\n    "))
}

type WC struct {
	State State

	ui   UI
	repo *Repository
}

func openWC(ui UI, repo *Repository) (*WC, error) {
	w := &WC{
		ui:   ui,
		repo: repo,
	}
	p := filepath.Join(repo.nzndir, "state.json")
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		if err := unmarshal(p, &w.State); err != nil {
			return nil, err
		}
	} else {
		w.State.WC = []*Entry{}
	}
	return w, nil
}

func (w *WC) Flush() error {
	return marshal(filepath.Join(w.repo.nzndir, "state.json"), &w.State)
}

func (w *WC) PathFor(path string) string {
	return filepath.Join(filepath.Dir(w.repo.nzndir), path)
}

func (w *WC) Exists(path string) bool {
	_, err := os.Lstat(w.PathFor(path))
	return !os.IsNotExist(err)
}

func (w *WC) IsLink(path string) bool {
	return isLink(filepath.Join(filepath.Dir(w.repo.nzndir), path))
}

func (w *WC) LinkTo(path string, layer *Layer) bool {
	for ; path != "."; path = filepath.Dir(path) {
		if linkTo(w.PathFor(path), filepath.Join(w.repo.repodir, layer.Path(), path)) {
			return true
		}
	}
	return false
}

func (w *WC) Link(layer *Layer, path string) error {
	dest := w.PathFor(path)
	for p, root := dest, filepath.Dir(w.repo.nzndir); p != root; {
		p = filepath.Dir(p)
		if isLink(p) {
			return &os.PathError{"link", p, errLink}
		}
	}
	if dir := filepath.Dir(dest); !w.Exists(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	return link(w.repo.PathFor(layer, path), dest)
}

func (w *WC) Unlink(path string) error {
	return unlink(path)
}

func (w *WC) SelectLayer(name string) error {
	l, err := w.repo.LayerOf(name)
	switch {
	case err != nil:
		return err
	case 0 < len(l.Layers):
		return fmt.Errorf("layer '%s' is abstract", name)
	case l.parent == nil:
		return fmt.Errorf("layer '%s' is not abstract", name)
	}
	for k, v := range w.State.Layers {
		if k == l.parent.Name {
			if v == l.Name {
				return fmt.Errorf("layer '%s' is already '%s'", k, v)
			}
			w.State.Layers[k] = l.Name
			return nil
		}
	}
	if w.State.Layers == nil {
		w.State.Layers = make(map[string]string)
	}
	w.State.Layers[l.parent.Name] = l.Name
	return nil
}

func (w *WC) LayerFor(name string) (*Layer, error) {
	for k, v := range w.State.Layers {
		if name == k {
			return w.repo.LayerOf(k + "/" + v)
		}
	}
	return nil, &ResolveError{Name: name}
}

func (w *WC) Layers() ([]*Layer, error) {
	list := make([]*Layer, len(w.repo.Layers))
	for i, l := range w.repo.Layers {
		if 0 < len(l.Layers) {
			wl, err := w.LayerFor(l.Name)
			if err == nil {
				list[i] = wl
				continue
			}
			list := make([]string, len(l.Layers))
			for i, sl := range l.Layers {
				list[i] = sl.Name
			}
			return nil, &ResolveError{l.Name, list}
		}
		list[i] = l
	}
	return list, nil
}

func (w *WC) MergeLayers() ([]*Entry, error) {
	layers, err := w.Layers()
	if err != nil {
		return nil, err
	}
	lwc := make(layeredWC)
	for _, l := range layers {
		err := w.repo.Walk(l.Path(), func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if _, ok := lwc[path]; !ok {
				for i := 0; i < len(path); i++ {
					if os.IsPathSeparator(path[i]) {
						lwc.add(path[:i], l, true)
					}
				}
				lwc.add(path, l, fi.IsDir())
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	wc := make(map[string]*Entry)
	for _, e := range w.State.WC {
		wc[e.Path] = e
	}
	w.State.WC = w.State.WC[:0]

	dir := ""
	for _, p := range w.sort(lwc) {
		switch {
		case dir != "" && strings.HasPrefix(p, dir):
		case len(lwc[p]) == 1:
			var l *Layer
			var b bool
			for l, b = range lwc[p] {
			}
			w.State.WC = append(w.State.WC, &Entry{l.Path(), p, b})
			if b {
				dir = p + "/"
			} else {
				dir = ""
			}
			if e, ok := wc[p]; ok {
				if e.Layer == l.Path() && e.IsDir == b {
					delete(wc, p)
				}
			}
		}
	}

	list := make([]*Entry, len(wc))
	i := 0
	for _, p := range w.sort(wc) {
		list[i] = wc[p]
		i++
	}
	return list, nil
}

func (w *WC) sort(m interface{}) []string {
	v := reflect.ValueOf(m)
	keys := v.MapKeys()
	list := make(sort.StringSlice, len(keys))
	i := 0
	for _, k := range keys {
		list[i] = k.String()
		i++
	}
	list.Sort()
	return list
}

func (w *WC) Errorf(err error) error {
	switch v := err.(type) {
	case *os.LinkError:
		if r, err := filepath.Rel(w.PathFor("."), v.New); err == nil {
			v.New = r
		}
		return fmt.Errorf("%s: %s", v.New, v.Err)
	case *os.PathError:
		if r, err := filepath.Rel(w.PathFor("."), v.Path); err == nil {
			v.Path = r
		}
		return fmt.Errorf("%s: %s", v.Path, v.Err)
	default:
		return err
	}
}

type layeredWC map[string]map[*Layer]bool

func (w layeredWC) add(p string, l *Layer, isDir bool) {
	if _, ok := w[p]; !ok {
		w[p] = make(map[*Layer]bool)
	}
	w[p][l] = isDir
}
