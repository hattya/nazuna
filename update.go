//
// nazuna :: update.go
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
	"fmt"
	"os"
)

var cmdUpdate = &Command{
	Names: []string{"update"},
	Usage: "update [<path>...]",
	Help: `
  update working copy
`,
}

func init() {
	cmdUpdate.Run = runUpdate
}

func runUpdate(ui UI, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	repo, err := OpenRepository(ui, wd)
	if err != nil {
		return err
	}

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
			return fmt.Errorf("%s: not tracked", e.Path)
		}
		ui.Println("unlink", entryPath(e), "-/-", e.Layer)
		switch l, err := repo.LayerOf(e.Layer); {
		case err != nil:
			return err
		case !wc.LinkTo(e.Path, l):
			return fmt.Errorf("not linked to layer '%s'", e.Layer)
		}
		if err := wc.Unlink(e.Path); err != nil {
			return err
		}
		removed++
	}

	for i := 0; i < len(wc.State.WC); i++ {
		e := wc.State.WC[i]
		l, _ := repo.LayerOf(e.Layer)
		if wc.LinkTo(e.Path, l) {
			continue
		}
		ui.Println("link", entryPath(e), "-->", l.Path())
		if err = wc.Link(l, e.Path); err != nil {
			ui.Errorln("error:", wc.Errorf(err))
			copy(wc.State.WC[i:], wc.State.WC[i+1:])
			wc.State.WC = wc.State.WC[:len(wc.State.WC)-1]
			failed++
			continue
		}
		updated++
	}

	ui.Printf("%d updated, %d removed, %d failed\n", updated, removed, failed)
	if err := wc.Flush(); err != nil {
		return err
	}
	if 0 < failed {
		return SystemExit(1)
	}
	return nil
}
