//
// nazuna :: nazuna.go
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
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
)

const Version = "0.1+"

type Layer struct {
	Name   string             `json:"name"`
	Layers []*Layer           `json:"layers,omitempty"`
	Links  map[string][]*Link `json:"links,omitempty"`

	abstract *Layer
}

func (l *Layer) Path() string {
	if l.abstract != nil {
		return l.abstract.Name + "/" + l.Name
	}
	return l.Name
}

type layerByName []*Layer

func (s layerByName) Len() int           { return len(s) }
func (s layerByName) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s layerByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type Link struct {
	Path []string `json:"path,omitempty"`
	Src  string   `json:"src"`
	Dst  string   `json:"dst"`
}

type linkByDst []*Link

func (s linkByDst) Len() int           { return len(s) }
func (s linkByDst) Less(i, j int) bool { return s[i].Dst < s[j].Dst }
func (s linkByDst) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type State struct {
	Layers map[string]string `json:"layers,omitempty"`
	WC     []*Entry          `json:"wc,omitempty"`
}

type Entry struct {
	Layer  string `json:"layer"`
	Path   string `json:"path"`
	Origin string `json:"origin,omitempty"`
	IsDir  bool   `json:"dir,omitempty"`
	Type   string `json:"type,omitempty"`
}

const unlinkableType = "_"

func (e *Entry) Format(format string) string {
	var sep, rhs string
	if e.IsDir {
		sep = "/"
	}
	if e.Origin == "" {
		rhs = e.Layer
	} else {
		rhs = filepath.FromSlash(e.Origin + sep)
	}
	return fmt.Sprintf(format, e.Path+sep, rhs)
}

type UI interface {
	Args() []string
	Print(...interface{}) (int, error)
	Printf(string, ...interface{}) (int, error)
	Println(...interface{}) (int, error)
	Error(...interface{}) (int, error)
	Errorf(string, ...interface{}) (int, error)
	Errorln(...interface{}) (int, error)
	Exec(*exec.Cmd) error
}

type SystemExit int

func (e SystemExit) Error() string {
	return fmt.Sprintf("exit status %d", e)
}

var (
	errArg     = errors.New("invalid arguments")
	errLink    = errors.New("file is a link")
	errNotLink = errors.New("not a link")
)
