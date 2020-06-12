//
// nazuna/cmd/nzn :: link.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
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
		Desc: strings.TrimSpace(cli.Dedent(`
			create a link for the specified path

			  link is used to create a link of <src> to <dst>, and will be managed by
			  update. If <src> is not found on update, it will be ignored without error.

			  The value of --path flag is a list of directories like PATH or GOPATH
			  environment variables, and it is used to search <src>.

			  You can refer environment variables in <path> and <src>. Supported formats
			  are ${var} and $var.
		`)),
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
		return cli.FlagError("--layer flag is required")
	default:
		if len(ctx.Args) != 2 {
			return cli.ErrArgs
		}
		l, err := repo.LayerOf(ctx.String("layer"))
		if err != nil {
			return err
		}
		dst, err := wc.Rel('.', ctx.Args[1])
		if err != nil {
			return err
		}
		if _, err = l.NewLink(filepath.SplitList(ctx.String("path")), ctx.Args[0], dst); err != nil {
			return err
		}
	}
	return repo.Flush()
}
