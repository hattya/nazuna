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
	dir  string
}

func openWC(ui UI, repo *Repository) (*WC, error) {
	w := &WC{
		ui:   ui,
		repo: repo,
		dir:  filepath.Dir(repo.nzndir),
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
	return filepath.Join(w.dir, path)
}

func (w *WC) Rel(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(w.dir, abs)
	if err != nil || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("'%s' is not under root", path)
	}
	return rel, nil
}

func (w *WC) Exists(path string) bool {
	_, err := os.Lstat(w.PathFor(path))
	return !os.IsNotExist(err)
}

func (w *WC) IsLink(path string) bool {
	return isLink(w.PathFor(path))
}

func (w *WC) LinksTo(path, src string) bool {
	return linksTo(w.PathFor(path), src)
}

func (w *WC) Link(src, dst string) error {
	dst = w.PathFor(dst)
	for p := filepath.Dir(dst); p != w.dir; p = filepath.Dir(p) {
		if isLink(p) {
			return &os.PathError{"link", p, errLink}
		}
	}
	if dir := filepath.Dir(dst); !w.Exists(dir) {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	return link(src, dst)
}

func (w *WC) Unlink(path string) error {
	path = w.PathFor(path)
	if err := unlink(path); err != nil {
		return err
	}
	for p := filepath.Dir(path); p != w.dir; p = filepath.Dir(p) {
		if isLink(p) || !isEmptyDir(p) {
			break
		}
		if err := os.Remove(p); err != nil {
			return err
		}
	}
	return nil
}

func (w *WC) SelectLayer(name string) error {
	l, err := w.repo.LayerOf(name)
	switch {
	case err != nil:
		return err
	case 0 < len(l.Layers):
		return fmt.Errorf("layer '%s' is abstract", name)
	case l.abstract == nil:
		return fmt.Errorf("layer '%s' is not abstract", name)
	}
	for k, v := range w.State.Layers {
		if k == l.abstract.Name {
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
	w.State.Layers[l.abstract.Name] = l.Name
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
			if err != nil {
				list := make([]string, len(l.Layers))
				for i, ll := range l.Layers {
					list[i] = ll.Name
				}
				return nil, &ResolveError{l.Name, list}
			}
			l = wl
		}
		list[i] = l
	}
	return list, nil
}

func (w *WC) MergeLayers() ([]*Entry, error) {
	b := wcBuilder{w: w}
	if err := b.build(); err != nil {
		return nil, err
	}

	wc := make(map[string]*Entry)
	for _, e := range w.State.WC {
		wc[e.Path] = e
	}
	w.State.WC = w.State.WC[:0]

	dir := ""
	for _, p := range w.sortKeys(b.wc) {
		switch {
		case dir != "" && strings.HasPrefix(p, dir):
		case len(b.wc[p]) == 1:
			e := b.wc[p][0]
			if e.Type == unlinkableType {
				continue
			}
			w.State.WC = append(w.State.WC, e)
			if e.IsDir {
				dir = p + "/"
			} else {
				dir = ""
			}
			if c, ok := wc[p]; ok {
				if c.Layer == e.Layer && c.IsDir == e.IsDir {
					delete(wc, p)
				}
			}
		}
	}

	ul := make([]*Entry, len(wc))
	for i, p := range w.sortKeys(wc) {
		ul[i] = wc[p]
	}
	return ul, nil
}

func (w *WC) sortKeys(m interface{}) []string {
	v := reflect.ValueOf(m)
	keys := v.MapKeys()
	list := make(sort.StringSlice, len(keys))
	for i, k := range keys {
		list[i] = k.String()
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

type wcBuilder struct {
	w     *WC
	wc    map[string][]*Entry
	layer string
	warn  map[string]bool
}

func (b *wcBuilder) build() error {
	layers, err := b.w.Layers()
	if err != nil {
		return err
	}
	b.wc = make(map[string][]*Entry)
	b.warn = make(map[string]bool)
	for _, l := range layers {
		b.layer = l.Path()
		if err := b.repo(); err != nil {
			return err
		}
		for dir, ll := range l.Links {
			for _, l := range ll {
				src := os.ExpandEnv(l.Src)
				dst := filepath.Join(dir, os.ExpandEnv(l.Dst))
				if 0 < len(l.Path) {
				loop:
					for _, v := range l.Path {
						for _, p := range filepath.SplitList(os.ExpandEnv(v)) {
							if b.link(filepath.Join(p, src), dst) {
								break loop
							}
						}
					}
				} else {
					b.link(src, dst)
				}
			}
		}
	}
	return nil
}

func (b *wcBuilder) repo() error {
	return b.w.repo.Walk(b.layer, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if _, ok := b.wc[path]; !ok {
			b.parentDirs(path, true)
			b.wc[path] = append(b.wc[path], &Entry{
				Layer: b.layer,
				Path:  path,
				IsDir: fi.IsDir(),
			})
		}
		return nil
	})
}

func (b *wcBuilder) link(src, dst string) bool {
	fi, err := os.Stat(src)
	if os.IsNotExist(err) {
		return false
	}
	if list, ok := b.wc[dst]; !ok {
		b.parentDirs(dst, false)
		b.wc[dst] = append(b.wc[dst], &Entry{
			Layer:  b.layer,
			Path:   dst,
			Origin: src,
			IsDir:  fi.IsDir(),
			Type:   "link",
		})
	} else if _, ok := b.warn[dst]; !ok && list[0].Type != "link" {
		b.w.ui.Errorf("warning: link: '%s' exists in the repository\n", dst)
		b.warn[dst] = true
	}
	return true
}

func (b *wcBuilder) parentDirs(path string, linkable bool) {
	find := func(p string) *Entry {
		for _, e := range b.wc[p] {
			if e.Layer == b.layer {
				return e
			}
		}
		return nil
	}
	for i, _ := range path {
		if os.IsPathSeparator(path[i]) {
			p := path[:i]
			e := find(p)
			if e == nil {
				e = &Entry{
					Layer: b.layer,
					Path:  p,
					IsDir: true,
				}
				b.wc[p] = append(b.wc[p], e)
			}
			if !linkable {
				e.Type = unlinkableType
			}
		}
	}
}
