//
// nazuna :: layer.go
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
)

var cmdLayer = &Command{
	Names: []string{"layer"},
	Usage: []string{
		"layer [<name>]",
		"layer -c <name>",
	},
	Help: `
  manage repository layers

options:

  -c, --create    create a new layer
`,
}

var layerCreate bool

func init() {
	cmdLayer.Run = runLayer
	cmdLayer.Flag.BoolVar(&layerCreate, "c", false, "")
	cmdLayer.Flag.BoolVar(&layerCreate, "create", false, "")
}

func runLayer(ui UI, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	repo, err := OpenRepository(ui, wd)
	if err != nil {
		return err
	}

	switch {
	case layerCreate:
		if len(args) != 1 {
			return ErrArg
		}
		if _, err := repo.NewLayer(args[0]); err != nil {
			return err
		}
		return repo.Flush()
	case 0 < len(args):
		if len(args) != 1 {
			return ErrArg
		}
		wc, err := repo.WC()
		if err != nil {
			return err
		}
		if err = wc.SelectLayer(args[0]); err != nil {
			return err
		}
		return wc.Flush()
	default:
		wc, err := repo.WC()
		if err != nil {
			return err
		}
		for _, l := range repo.Layers {
			ui.Println(l.Name)
			for _, ll := range l.Layers {
				var s string
				if wl, err := wc.LayerFor(l.Name); err == nil && wl.Name == ll.Name {
					s = "*"
				}
				ui.Printf("    %s%s\n", ll.Name, s)
			}
		}
		return nil
	}
}
