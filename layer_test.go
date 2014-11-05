//
// nazuna :: layer_test.go
//
//   Copyright (c) 2014 Akinori Hattori <hattya@gmail.com>
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

package nazuna_test

import (
	"testing"

	"github.com/hattya/nazuna"
)

func TestLayer(t *testing.T) {
	l := &nazuna.Layer{Name: "layer"}
	if g, e := l.Path(), "layer"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}

	l.Abst(&nazuna.Layer{Name: "abst"})
	if g, e := l.Path(), "abst/layer"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
}

func TestSortLayers(t *testing.T) {
	layers := []*nazuna.Layer{
		{Name: "b"},
		{Name: "a"},
	}
	nazuna.SortLayers(layers)
	if g, e := layers[0].Name, "a"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := layers[1].Name, "b"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
}

func TestSortLinks(t *testing.T) {
	links := []*nazuna.Link{
		{Dst: "b"},
		{Dst: "a"},
	}
	nazuna.LinkSlice(links).Sort()
	if g, e := links[0].Dst, "a"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
	if g, e := links[1].Dst, "b"; g != e {
		t.Errorf("expected %v, got %v", e, g)
	}
}
