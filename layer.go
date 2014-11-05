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

import "sort"

type Layer struct {
	Name     string                `json:"name"`
	Layers   []*Layer              `json:"layers,omitempty"`
	Aliases  map[string]string     `json:"aliases,omitempty"`
	Links    map[string][]*Link    `json:"links,omitempty"`
	Subrepos map[string][]*Subrepo `json:"subrepos,omitempty"`

	abst *Layer
}

func (l *Layer) Path() string {
	if l.abst != nil {
		return l.abst.Name + "/" + l.Name
	}
	return l.Name
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

type LinkSlice []*Link

func (p LinkSlice) Len() int           { return len(p) }
func (p LinkSlice) Less(i, j int) bool { return p[i].Dst < p[j].Dst }
func (p LinkSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p LinkSlice) Sort() { sort.Sort(p) }
