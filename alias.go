//
// nazuna :: alias.go
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

package nazuna

import (
	"fmt"
	"os"
)

var cmdAlias = &Command{
	Names: []string{"alias"},
	Usage: []string{
		"alias -l <layer> <src> <dst>",
	},
	Help: `
create an alias for the specified path

  Change the location of <src> to <dst>. <src> should be existed in the lower layer
  than <dst>, and <src> is treated as <dst> in the layer <layer>. If <src> does
  not match any locations on update, it will be ignored without error.

  You can refer environment variables in <dst>. Supported formats are ${var}
  and $var.

options:

  -l, --layer    a layer
`,
}

var aliasL string

func init() {
	cmdAlias.Flag.StringVar(&aliasL, "l", "", "")
	cmdAlias.Flag.StringVar(&aliasL, "layer", "", "")

	cmdAlias.Run = runAlias
}

func runAlias(ui UI, args []string) error {
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

	switch {
	case aliasL == "":
		return FlagError("flag --layer is required")
	default:
		if len(args) != 2 {
			return ErrArg
		}
		l, err := repo.LayerOf(aliasL)
		switch {
		case err != nil:
			return err
		case 0 < len(l.Layers):
			return fmt.Errorf("layer '%s' is abstract", l.Path())
		}
		src, err := wc.Rel('/', args[0])
		if err != nil {
			return err
		}
		dst, err := wc.Rel('.', args[1])
		switch {
		case err != nil:
			return err
		case src == dst:
			return fmt.Errorf("'%s' and '%s' are the same file", src, dst)
		}
		switch typ := repo.Find(l, dst); typ {
		case "", "dir":
		case "file":
			return fmt.Errorf("'%s' already exists!", dst)
		default:
			return fmt.Errorf("%s '%s' already exists!", typ, dst)
		}
		if l.Aliases == nil {
			l.Aliases = make(map[string]string)
		}
		l.Aliases[src] = dst
	}
	return repo.Flush()
}
