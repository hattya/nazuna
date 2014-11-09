//
// nazuna :: layer.go
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
	"fmt"
	"path/filepath"
	"sort"
)

type Layer struct {
	Name     string                `json:"name"`
	Layers   []*Layer              `json:"layers,omitempty"`
	Aliases  map[string]string     `json:"aliases,omitempty"`
	Links    map[string][]*Link    `json:"links,omitempty"`
	Subrepos map[string][]*Subrepo `json:"subrepos,omitempty"`

	repo *Repository
	abst *Layer
}

func (l *Layer) Path() string {
	if l.abst != nil {
		return l.abst.Name + "/" + l.Name
	}
	return l.Name
}

func (l *Layer) NewAlias(src, dst string) error {
	if src == dst {
		return fmt.Errorf("'%v' and '%v' are the same path", src, dst)
	}
	if err := l.check(dst, true); err != nil {
		return err
	}

	if l.Aliases == nil {
		l.Aliases = make(map[string]string)
	}
	l.Aliases[src] = dst
	return nil
}

func (l *Layer) NewLink(path []string, src, dst string) (*Link, error) {
	if err := l.check(dst, false); err != nil {
		return nil, err
	}

	for i, p := range path {
		path[i] = filepath.ToSlash(filepath.Clean(p))
	}
	src = filepath.ToSlash(filepath.Clean(src))
	dir, dst := SplitPath(dst)
	lnk := &Link{
		Path: path,
		Src:  src,
		Dst:  dst,
	}
	if l.Links == nil {
		l.Links = make(map[string][]*Link)
	}
	l.Links[dir] = append(l.Links[dir], lnk)
	linkSlice(l.Links[dir]).Sort()
	return lnk, nil
}

func (l *Layer) NewSubrepo(src, dst string) (*Subrepo, error) {
	if err := l.check(dst, false); err != nil {
		return nil, err
	}

	dir, name := SplitPath(dst)
	if name == filepath.Base(src) {
		name = ""
	}
	sub := &Subrepo{
		Src:  src,
		Name: name,
	}
	if l.Subrepos == nil {
		l.Subrepos = make(map[string][]*Subrepo)
	}
	l.Subrepos[dir] = append(l.Subrepos[dir], sub)
	subrepoSlice(l.Subrepos[dir]).Sort()
	return sub, nil
}

func (l *Layer) check(path string, dir bool) error {
	if 0 < len(l.Layers) {
		return fmt.Errorf("layer '%v' is abstract", l.Path())
	}
	switch typ := l.repo.Find(l, path); typ {
	case "":
	case "file", "dir":
		if dir && typ == "dir" {
			break
		}
		return fmt.Errorf("'%v' already exists!", path)
	default:
		return fmt.Errorf("%v '%v' already exists!", typ, path)
	}
	return nil
}

type layerSlice []*Layer

func (p layerSlice) Len() int           { return len(p) }
func (p layerSlice) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p layerSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p layerSlice) Sort() { sort.Sort(p) }

type Link struct {
	Path []string `json:"path,omitempty"`
	Src  string   `json:"src"`
	Dst  string   `json:"dst"`
}

type linkSlice []*Link

func (p linkSlice) Len() int           { return len(p) }
func (p linkSlice) Less(i, j int) bool { return p[i].Dst < p[j].Dst }
func (p linkSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p linkSlice) Sort() { sort.Sort(p) }

type Subrepo struct {
	Src  string `json:"src"`
	Name string `json:"name,omitempty"`
}

type subrepoSlice []*Subrepo

func (p subrepoSlice) Len() int           { return len(p) }
func (p subrepoSlice) Less(i, j int) bool { return p[i].Src < p[j].Src }
func (p subrepoSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p subrepoSlice) Sort() { sort.Sort(p) }
