//
// nazuna :: wc.go
//
//   Copyright (c) 2013-2017 Akinori Hattori <hattya@gmail.com>
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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrLink    = errors.New("file is a link")
	ErrNotLink = errors.New("not a link")
)

type ResolveError struct {
	Name string
	List []string
}

func (e *ResolveError) Error() string {
	s := fmt.Sprintf("cannot resolve layer '%v'", e.Name)
	if len(e.List) == 0 {
		return s
	}
	return fmt.Sprintf("%v:\n    %v", s, strings.Join(e.List, "\n    "))
}

type WC struct {
	State State

	ui   UI
	repo *Repository
}

func openWC(ui UI, repo *Repository) (*WC, error) {
	wc := &WC{
		ui:   ui,
		repo: repo,
	}
	if err := unmarshal(repo, filepath.Join(repo.nzndir, "state.json"), &wc.State); err != nil {
		return nil, err
	}
	if wc.State.WC == nil {
		wc.State.WC = []*Entry{}
	}
	return wc, nil
}

func (wc *WC) Flush() error {
	return marshal(wc.repo, filepath.Join(wc.repo.nzndir, "state.json"), &wc.State)
}

func (wc *WC) PathFor(path string) string {
	return filepath.Join(wc.repo.root, path)
}

func (wc *WC) Rel(base rune, path string) (string, error) {
	if strings.HasPrefix(path, "$") {
		return filepath.ToSlash(path), nil
	}

	var abs string
	var err error
	switch base {
	case '/':
		if filepath.IsAbs(path) {
			abs = path
		} else {
			abs = filepath.Join(wc.repo.root, path)
		}
	case '.':
		if abs, err = filepath.Abs(path); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unknown base '%c'", base)
	}
	rel, err := filepath.Rel(wc.repo.root, abs)
	if err != nil || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("'%v' is not under root", path)
	}
	return filepath.ToSlash(rel), nil
}

func (wc *WC) Exists(path string) bool {
	_, err := os.Lstat(wc.PathFor(path))
	return err == nil
}

func (wc *WC) IsLink(path string) bool {
	return IsLink(wc.PathFor(path))
}

func (wc *WC) LinksTo(path, origin string) bool {
	return LinksTo(wc.PathFor(path), origin)
}

func (wc *WC) Link(src, dst string) error {
	dst = wc.PathFor(dst)
	for p := filepath.Dir(dst); p != wc.repo.root; p = filepath.Dir(p) {
		if IsLink(p) {
			return &os.PathError{
				Op:   "link",
				Path: p,
				Err:  ErrLink,
			}
		}
	}
	dir := filepath.Dir(dst)
	if _, err := os.Lstat(dir); err != nil {
		if err := os.MkdirAll(dir, 0777); err != nil {
			return err
		}
	}
	return CreateLink(src, dst)
}

func (wc *WC) Unlink(path string) error {
	path = wc.PathFor(path)
	if err := Unlink(path); err != nil {
		return err
	}
	for p := filepath.Dir(path); p != wc.repo.root; p = filepath.Dir(p) {
		if IsLink(p) || !IsEmptyDir(p) {
			break
		}
		if err := os.Remove(p); err != nil {
			return err
		}
	}
	return nil
}

func (wc *WC) SelectLayer(name string) error {
	l, err := wc.repo.LayerOf(name)
	switch {
	case err != nil:
		return err
	case 0 < len(l.Layers):
		return fmt.Errorf("layer '%v' is abstract", name)
	case l.abst == nil:
		return fmt.Errorf("layer '%v' is not abstract", name)
	}
	for k, v := range wc.State.Layers {
		if k == l.abst.Name {
			if v == l.Name {
				return fmt.Errorf("layer '%v' is already '%v'", k, v)
			}
			wc.State.Layers[k] = l.Name
			return nil
		}
	}
	if wc.State.Layers == nil {
		wc.State.Layers = make(map[string]string)
	}
	wc.State.Layers[l.abst.Name] = l.Name
	return nil
}

func (wc *WC) LayerFor(name string) (*Layer, error) {
	for k, v := range wc.State.Layers {
		if name == k {
			return wc.repo.LayerOf(k + "/" + v)
		}
	}
	return nil, &ResolveError{Name: name}
}

func (wc *WC) Layers() ([]*Layer, error) {
	list := make([]*Layer, len(wc.repo.Layers))
	for i, l := range wc.repo.Layers {
		if 0 < len(l.Layers) {
			wl, err := wc.LayerFor(l.Name)
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

func (wc *WC) MergeLayers() ([]*Entry, error) {
	b := wcBuilder{w: wc}
	if err := b.build(); err != nil {
		return nil, err
	}

	wc.State.WC = wc.State.WC[:0]
	dir := ""
	for _, p := range sortKeys(b.wc) {
		switch {
		case dir != "" && strings.HasPrefix(p, dir):
		case len(b.wc[p]) == 1:
			e := b.wc[p][0]
			if e.Type == unlinkableType {
				continue
			}
			wc.State.WC = append(wc.State.WC, e)
			if e.IsDir {
				dir = p + "/"
			} else {
				dir = ""
			}
			if c, ok := b.state[p]; ok {
				if c.Layer == e.Layer && c.IsDir == e.IsDir {
					delete(b.state, p)
				}
			}
		}
	}

	ul := make([]*Entry, len(b.state))
	for i, p := range sortKeys(b.state) {
		ul[i] = b.state[p]
	}
	return ul, nil
}

func (wc *WC) Errorf(err error) error {
	switch v := err.(type) {
	case *os.LinkError:
		if r, err := wc.Rel('/', v.New); err == nil {
			v.New = filepath.ToSlash(r)
		}
		return fmt.Errorf("%v: %v", v.New, v.Err)
	case *os.PathError:
		if r, err := wc.Rel('/', v.Path); err == nil {
			v.Path = filepath.ToSlash(r)
		}
		return fmt.Errorf("%v: %v", v.Path, v.Err)
	}
	return err
}

type wcBuilder struct {
	w       *WC
	l       *Layer
	layer   string
	state   map[string]*Entry
	wc      map[string][]*Entry
	aliases map[string]string
}

func (b *wcBuilder) build() error {
	layers, err := b.w.Layers()
	if err != nil {
		return err
	}
	b.state = make(map[string]*Entry)
	for _, e := range b.w.State.WC {
		b.state[e.Path] = e
	}
	b.wc = make(map[string][]*Entry)
	b.aliases = make(map[string]string)
	for _, l := range layers {
		b.l = l
		b.layer = l.Path()
		if err := b.repo(); err != nil {
			return err
		}
		if err := b.link(); err != nil {
			return err
		}
		if err := b.subrepo(); err != nil {
			return err
		}
		for src, dst := range b.l.Aliases {
			if _, ok := b.aliases[src]; !ok {
				b.aliases[src] = dst
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
		origin := path[len(b.layer)+1:]
		path, err = b.alias(origin)
		if err != nil {
			return err
		}
		if _, ok := b.wc[path]; !ok {
			b.parentDirs(path, true)
			e := &Entry{
				Layer: b.layer,
				Path:  path,
				IsDir: fi.IsDir(),
			}
			b.wc[path] = append(b.wc[path], e)
			if path != origin {
				e.Origin = origin
				for p, o := filepath.Dir(path), filepath.Dir(origin); p != "."; p = filepath.Dir(p) {
					e := b.find(filepath.ToSlash(p))
					if o != "." {
						e.Origin = filepath.ToSlash(o)
						o = filepath.Dir(o)
					} else {
						e.Type = unlinkableType
					}
				}
			}
		}
		return nil
	})
}

func (b *wcBuilder) link() error {
	link := func(src, dst string) (bool, error) {
		fi, err := os.Stat(src)
		if err != nil {
			return false, nil
		}
		dst, err = b.alias(dst)
		if err != nil {
			return false, fmt.Errorf("link %v", err)
		}
		switch list, ok := b.wc[dst]; {
		case !ok:
			b.parentDirs(dst, false)
			b.wc[dst] = append(b.wc[dst], &Entry{
				Layer:  b.layer,
				Path:   dst,
				Origin: src,
				IsDir:  fi.IsDir(),
				Type:   "link",
			})
		case list[0].Layer == b.layer && list[0].Type != "link":
			b.w.ui.Errorf("warning: link: '%v' exists in the repository\n", dst)
		}
		return true, nil
	}
	for _, dir := range sortKeys(b.l.Links) {
		for _, l := range b.l.Links[dir] {
			src := filepath.FromSlash(filepath.Clean(os.ExpandEnv(l.Src)))
			dst := filepath.ToSlash(filepath.Join(dir, l.Dst))
			if 0 < len(l.Path) {
			L:
				for _, v := range l.Path {
					for _, p := range filepath.SplitList(os.ExpandEnv(v)) {
						switch ok, err := link(filepath.FromSlash(filepath.Clean(filepath.Join(p, src))), dst); {
						case ok:
							break L
						case err != nil:
							return err
						}
					}
				}
			} else if _, err := link(src, dst); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *wcBuilder) subrepo() error {
	for _, dir := range sortKeys(b.l.Subrepos) {
		for _, sub := range b.l.Subrepos[dir] {
			var name string
			if sub.Name != "" {
				name = sub.Name
			} else {
				name = filepath.Base(sub.Src)
			}
			dst, err := b.alias(filepath.ToSlash(filepath.Join(dir, name)))
			if err != nil {
				return fmt.Errorf("subrepo %v", err)
			}
			switch list, ok := b.wc[dst]; {
			case !ok:
				b.parentDirs(dst, false)
				b.wc[dst] = append(b.wc[dst], &Entry{
					Layer:  b.layer,
					Path:   dst,
					Origin: sub.Src,
					Type:   "subrepo",
				})
			case list[0].Layer == b.layer && list[0].Type != "subrepo":
				b.w.ui.Errorf("warning: subrepo: '%v' exists in the repository\n", dst)
			}
		}
	}
	return nil
}

func (b *wcBuilder) alias(path string) (string, error) {
	for src := path; src != "."; src = filepath.Dir(src) {
		if dst, ok := b.aliases[src]; ok {
			if path == src {
				path = dst
			} else {
				path = filepath.Join(dst, path[len(src)+1:])
			}
			return b.w.Rel('/', filepath.Clean(os.ExpandEnv(path)))
		}
	}
	return path, nil
}

func (b *wcBuilder) parentDirs(path string, linkable bool) {
	inWC := true
	for i, _ := range path {
		if path[i] == '/' {
			p := path[:i]
			e := b.find(p)
			if e == nil {
				e = &Entry{
					Layer: b.layer,
					Path:  p,
					IsDir: true,
				}
				b.wc[p] = append(b.wc[p], e)
			}
			switch {
			case !inWC:
			case linkable:
				switch _, ok := b.state[p]; {
				case ok:
					inWC = false
					if b.w.Exists(p) && !b.w.IsLink(p) {
						e.Type = unlinkableType
						delete(b.state, p)
					}
				case b.w.Exists(p):
					e.Type = unlinkableType
				}
			case !linkable:
				e.Type = unlinkableType
			}
		}
	}
}

func (b *wcBuilder) find(p string) *Entry {
	for _, e := range b.wc[p] {
		if e.Layer == b.layer {
			return e
		}
	}
	return nil
}
