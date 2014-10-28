//
// nzn :: subrepo.go
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
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hattya/go.cli"
	"github.com/hattya/nazuna"
)

func init() {
	flags := cli.NewFlagSet()
	flags.String("l, layer", "", "layer name")
	flags.Bool("a, add", false, "add <repository> to <path>")
	flags.Bool("u, update", false, "clone or update repositories")

	app.Add(&cli.Command{
		Name: []string{"subrepo"},
		Usage: []string{
			"-l <layer> -a <repository> <path>",
			"-u",
		},
		Desc: strings.TrimSpace(`
manage subrepositories

  subrepo is used to manage external repositories.

  subrepo can associate <repository> to <path> by flag --add. If <path> ends
  with a path separator, it will be associated as the basename of <repository>
  under <path>.

  subrepo can clone or update the repositories in the working copy by flag
  --update.
`),
		Flags:  flags,
		Action: subrepo,
		Data:   true,
	})
}

func subrepo(ctx *cli.Context) error {
	repo := ctx.Data.(*nazuna.Repository)
	wc, err := repo.WC()
	if err != nil {
		return err
	}

	switch {
	case ctx.Bool("add"):
		switch {
		case ctx.String("layer") == "":
			return cli.FlagError("flag --layer is required")
		case len(ctx.Args) != 2:
			return cli.ErrArgs
		}
		l, err := repo.LayerOf(ctx.String("layer"))
		switch {
		case err != nil:
			return err
		case 0 < len(l.Layers):
			return fmt.Errorf("layer '%v' is abstract", l.Path())
		}
		src := ctx.Args[0]
		path, err := wc.Rel('.', ctx.Args[1])
		if err != nil {
			return err
		}
		var name, dst string
		if 0 < len(ctx.Args[1]) && os.IsPathSeparator(ctx.Args[1][len(ctx.Args[1])-1]) {
			dst = path + "/" + filepath.Base(src)
		} else {
			path, name = nazuna.SplitPath(path)
			dst = path + "/" + name
			if name == filepath.Base(src) {
				name = ""
			}
		}
		switch typ := repo.Find(l, dst); typ {
		case "":
		case "dir", "file":
			return fmt.Errorf("'%v' already exists!", dst)
		default:
			return fmt.Errorf("%v '%v' already exists!", typ, dst)
		}
		if l.Subrepos == nil {
			l.Subrepos = make(map[string][]*nazuna.Subrepo)
		}
		l.Subrepos[path] = append(l.Subrepos[path], &nazuna.Subrepo{
			Src:  src,
			Name: name,
		})
		sort.Sort(nazuna.SubrepoBySrc(l.Subrepos[path]))
		return repo.Flush()
	case ctx.Bool("update"):
		_, err := wc.MergeLayers()
		if err != nil {
			return err
		}
		ui := newUI()
		for _, e := range wc.State.WC {
			if e.Type != "subrepo" {
				continue
			}
			app.Printf("* %v\n", e.Origin)
			r, err := nazuna.NewRemote(ui, e.Origin)
			if err != nil {
				return err
			}
			dst := repo.SubrepoFor(r.Root)
			if nazuna.IsEmptyDir(dst) {
				dst, _ = wc.Rel('.', dst)
				err = r.Clone(wc.PathFor("/"), dst)
			} else {
				err = r.Update(dst)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
