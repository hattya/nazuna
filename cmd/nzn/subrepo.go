//
// nazuna/cmd/nzn :: subrepo.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"os"
	"path/filepath"
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
		Desc: strings.TrimSpace(cli.Dedent(`
			manage subrepositories

			  subrepo is used to manage external repositories.

			  subrepo can associate <repository> to <path> by --add flag. If <path> ends
			  with a path separator, it will be associated as the basename of <repository>
			  under <path>.

			  subrepo can clone or update the repositories in the working copy by --update
			  flag.
		`)),
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
			return cli.FlagError("--layer flag is required")
		case len(ctx.Args) != 2:
			return cli.ErrArgs
		}
		l, err := repo.LayerOf(ctx.String("layer"))
		if err != nil {
			return err
		}
		src := ctx.Args[0]
		dst := ctx.Args[1]
		rel, err := wc.Rel('.', dst)
		if err != nil {
			return err
		}
		if 0 < len(dst) && os.IsPathSeparator(dst[len(dst)-1]) {
			dst = rel + "/" + filepath.Base(src)
		} else {
			dst = rel
		}
		if _, err := l.NewSubrepo(src, dst); err != nil {
			return err
		}
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
