//
// nzn :: alias.go
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
	"fmt"
	"strings"

	"github.com/hattya/go.cli"
	"github.com/hattya/nazuna"
)

func init() {
	flags := cli.NewFlagSet()
	flags.String("l, layer", "", "layer name")

	app.Add(&cli.Command{
		Name:  []string{"alias"},
		Usage: "-l <layer> <src> <dst>",
		Desc: strings.TrimSpace(`
create an alias for the specified path

  Change the location of <src> to <dst>. <src> should be existed in the lower layer
  than <dst>, and <src> is treated as <dst> in the layer <layer>. If <src> does
  not match any locations on update, it will be ignored without error.

  You can refer environment variables in <dst>. Supported formats are ${var}
  and $var.
`),
		Flags:  flags,
		Action: alias,
		Data:   true,
	})
}

func alias(ctx *cli.Context) error {
	repo := ctx.Data.(*nazuna.Repository)
	wc, err := repo.WC()
	if err != nil {
		return err
	}

	switch {
	case ctx.String("layer") == "":
		return cli.FlagError("flag --layer is required")
	default:
		if len(ctx.Args) != 2 {
			return cli.ErrArgs
		}
		l, err := repo.LayerOf(ctx.String("layer"))
		if err != nil {
			return err
		}
		src, err := wc.Rel('/', ctx.Args[0])
		if err != nil {
			return err
		}
		dst, err := wc.Rel('.', ctx.Args[1])
		if err != nil {
			return err
		}
		switch typ := repo.Find(l, dst); typ {
		case "", "dir":
		case "file":
			return fmt.Errorf("'%v' already exists!", dst)
		default:
			return fmt.Errorf("%v '%v' already exists!", typ, dst)
		}
		if err := l.NewAlias(src, dst); err != nil {
			return err
		}
	}
	return repo.Flush()
}
