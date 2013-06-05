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
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

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
		if linkTo(w.PathFor(path), filepath.Join(w.repo.repodir, layer.Name, path)) {
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

func (w *WC) MergeLayers() ([]*Entry, error) {
	lwc := make(layeredWC)
	for _, l := range w.repo.Layers {
		err := w.repo.Walk(l.Name, func(path string, fi os.FileInfo, err error) error {
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
			w.State.WC = append(w.State.WC, &Entry{l.Name, p, b})
			if b {
				dir = p + "/"
			} else {
				dir = ""
			}
			if e, ok := wc[p]; ok {
				if e.Layer == l.Name && e.IsDir == b {
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

type layeredWC map[string]map[*Layer]bool

func (w layeredWC) add(p string, l *Layer, isDir bool) {
	if _, ok := w[p]; !ok {
		w[p] = make(map[*Layer]bool)
	}
	w[p][l] = isDir
}
