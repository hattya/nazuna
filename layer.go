//
// nazuna :: layer.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
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
	s := filepath.ToSlash(filepath.Clean(src))
	d := filepath.ToSlash(filepath.Clean(dst))
	if s == d {
		return fmt.Errorf("'%v' and '%v' are the same path", src, dst)
	}
	if err := l.check(dst, true); err != nil {
		return err
	}

	if l.Aliases == nil {
		l.Aliases = make(map[string]string)
	}
	l.Aliases[s] = d
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
	dir, dst := SplitPath(filepath.ToSlash(filepath.Clean(dst)))
	lnk := &Link{
		Path: path,
		Src:  src,
		Dst:  dst,
	}
	if l.Links == nil {
		l.Links = make(map[string][]*Link)
	}
	l.Links[dir] = append(l.Links[dir], lnk)
	sort.Slice(l.Links[dir], func(i, j int) bool { return l.Links[dir][i].Dst < l.Links[dir][j].Dst })
	return lnk, nil
}

func (l *Layer) NewSubrepo(src, dst string) (*Subrepo, error) {
	if err := l.check(dst, false); err != nil {
		return nil, err
	}

	dir, name := SplitPath(filepath.ToSlash(filepath.Clean(dst)))
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
	sort.Slice(l.Subrepos[dir], func(i, j int) bool { return l.Subrepos[dir][i].Src < l.Subrepos[dir][j].Src })
	return sub, nil
}

func (l *Layer) check(path string, dir bool) error {
	if len(l.Layers) != 0 {
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

type Link struct {
	Path []string `json:"path,omitempty"`
	Src  string   `json:"src"`
	Dst  string   `json:"dst"`
}

type Subrepo struct {
	Src  string `json:"src"`
	Name string `json:"name,omitempty"`
}
