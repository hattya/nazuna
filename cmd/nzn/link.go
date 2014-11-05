//
// nzn :: link.go
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
	"path/filepath"
	"strings"

	"github.com/hattya/go.cli"
	"github.com/hattya/nazuna"
)

func init() {
	flags := cli.NewFlagSet()
	flags.String("l, layer", "", "layer name")
	flags.String("p, path", "", "list of directories to search <src>")

	app.Add(&cli.Command{
		Name:  []string{"link"},
		Usage: "-l <layer> [-p <path>] <src> <dst>",
		Desc: strings.TrimSpace(`
create a link for the specified path

  link is used to create a link of <src> to <dst>, and will be managed by
  update. If <src> is not found on update, it will be ignored without error.

  The value of flag --path is a list of directories like PATH or GOPATH
  environment variables, and it is used to search <src>.

  You can refer environment variables in <path> and <src>. Supported formats
  are ${var} and $var.
`),
		Flags:  flags,
		Action: link,
		Data:   true,
	})
}

func link(ctx *cli.Context) error {
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
		switch {
		case err != nil:
			return err
		case 0 < len(l.Layers):
			return fmt.Errorf("layer '%v' is abstract", l.Path())
		}
		dst, err := wc.Rel('.', ctx.Args[1])
		if err != nil {
			return err
		}
		switch typ := repo.Find(l, dst); typ {
		case "":
		case "dir", "file":
			return fmt.Errorf("'%v' already exists!", dst)
		default:
			return fmt.Errorf("%v '%v' already exists!", typ, dst)
		}
		path := filepath.SplitList(ctx.String("path"))
		for i, p := range path {
			path[i] = filepath.ToSlash(filepath.Clean(p))
		}
		src := filepath.ToSlash(filepath.Clean(ctx.Args[0]))
		dir, dst := nazuna.SplitPath(dst)
		if l.Links == nil {
			l.Links = make(map[string][]*nazuna.Link)
		}
		l.Links[dir] = append(l.Links[dir], &nazuna.Link{
			Path: path,
			Src:  src,
			Dst:  dst,
		})
		nazuna.LinkSlice(l.Links[dir]).Sort()
	}
	return repo.Flush()
}
