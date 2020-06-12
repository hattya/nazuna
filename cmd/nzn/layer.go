//
// nazuna/cmd/nzn :: layer.go
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
	flags.Bool("c, create", false, "create a new layer")

	app.Add(&cli.Command{
		Name: []string{"layer"},
		Usage: []string{
			"[<name>]",
			"-c <name>",
		},
		Desc: strings.TrimSpace(cli.Dedent(`
			manage repository layers
		`)),
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
