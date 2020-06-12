//
// nazuna/cmd/nzn :: alias.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
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
		Desc: strings.TrimSpace(cli.Dedent(`
			create an alias for the specified path

			  Change the location of <src> to <dst>. <src> should be existed in the lower layer
			  than <dst>, and <src> is treated as <dst> in the layer <layer>. If <src> does
			  not match any locations on update, it will be ignored without error.

			  You can refer environment variables in <dst>. Supported formats are ${var}
			  and $var.
		`)),
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
		return cli.FlagError("--layer flag is required")
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
		if err := l.NewAlias(src, dst); err != nil {
			return err
		}
	}
	return repo.Flush()
}
