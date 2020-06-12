//
// nazuna/cmd/nzn :: update.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
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

	app.Add(&cli.Command{
		Name:  []string{"update"},
		Usage: "update",
		Desc: strings.TrimSpace(cli.Dedent(`
			update working copy

			  Update links in the working copy to match with the repository configuration.
		`)),
		Flags:  flags,
		Action: update,
		Data:   true,
	})
}

func update(ctx *cli.Context) error {
	repo := ctx.Data.(*nazuna.Repository)
	wc, err := repo.WC()
	if err != nil {
		return err
	}
	ul, err := wc.MergeLayers()
	if err != nil {
		return wc.Errorf(err)
	}

	updated, removed, failed := 0, 0, 0
	for _, e := range ul {
		switch {
		case !wc.Exists(e.Path):
			continue
		case !wc.IsLink(e.Path):
			return fmt.Errorf("%v: not tracked", e.Path)
		}
		app.Println(e.Format("unlink %v -/- %v"))
		switch e.Type {
		case "link":
			if !wc.LinksTo(e.Path, e.Origin) {
				return fmt.Errorf("not linked to '%v'", e.Origin)
			}
		case "subrepo":
			if !wc.LinksTo(e.Path, repo.SubrepoFor(e.Origin)) {
				return fmt.Errorf("not linked to '%v'", e.Origin)
			}
		default:
			var origin string
			if e.Origin != "" {
				origin = e.Origin
			} else {
				origin = e.Path
			}
			if !wc.LinksTo(e.Path, repo.PathFor(nil, filepath.Join(e.Layer, origin))) {
				return fmt.Errorf("not linked to layer '%v'", e.Layer)
			}
		}
		if err := wc.Unlink(e.Path); err != nil {
			return err
		}
		removed++
	}

	for i := 0; i < len(wc.State.WC); i++ {
		e := wc.State.WC[i]
		var origin string
		switch e.Type {
		case "link":
			origin = e.Origin
		case "subrepo":
			origin = repo.SubrepoFor(e.Origin)
			if !nazuna.IsDir(origin) {
				continue
			}
		default:
			l, _ := repo.LayerOf(e.Layer)
			if e.Origin != "" {
				origin = repo.PathFor(l, e.Origin)
			} else {
				origin = repo.PathFor(l, e.Path)
			}
		}
		if wc.LinksTo(e.Path, origin) {
			continue
		}
		app.Println(e.Format("link %v --> %v"))
		if err := wc.Link(origin, e.Path); err != nil {
			app.Errorln("error:", wc.Errorf(err))
			copy(wc.State.WC[i:], wc.State.WC[i+1:])
			wc.State.WC = wc.State.WC[:len(wc.State.WC)-1]
			i--
			failed++
		} else {
			updated++
		}
	}

	app.Printf("%d updated, %d removed, %d failed\n", updated, removed, failed)
	if err := wc.Flush(); err != nil {
		return err
	}
	if 0 < failed {
		return SystemExit(1)
	}
	return nil
}
