//
// nzn :: layer.go
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

package main

import (
	"strings"

	"github.com/hattya/go.cli"
	"github.com/hattya/nazuna"
)

func init() {
	flags := cli.NewFlagSet()
	flags.Bool("c, create", false, "create a new layer")

	app.Add(&cli.Command{
		Name: []string{"layer"},
		Usage: []string{
			"[<name>]",
			"-c <name>",
		},
		Desc: strings.TrimSpace(`
manage repository layers
`),
		Flags:  flags,
		Action: layer,
		Data:   true,
	})
}

func layer(ctx *cli.Context) error {
	repo := ctx.Data.(*nazuna.Repository)
	switch {
	case ctx.Bool("create"):
		if len(ctx.Args) != 1 {
			return cli.ErrArgs
		}
		if _, err := repo.NewLayer(ctx.Args[0]); err != nil {
			return err
		}
		return repo.Flush()
	case 0 < len(ctx.Args):
		if len(ctx.Args) != 1 {
			return cli.ErrArgs
		}
		wc, err := repo.WC()
		if err != nil {
			return err
		}
		if err := wc.SelectLayer(ctx.Args[0]); err != nil {
			return err
		}
		return wc.Flush()
	default:
		wc, err := repo.WC()
		if err != nil {
			return err
		}
		for _, l := range repo.Layers {
			app.Println(l.Name)
			for _, ll := range l.Layers {
				var s string
				if wl, err := wc.LayerFor(l.Name); err == nil && wl.Name == ll.Name {
					s = "*"
				}
				app.Printf("    %v%v\n", ll.Name, s)
			}
		}
		return nil
	}
}
