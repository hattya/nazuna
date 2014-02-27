//
// nazuna :: link.go
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
	"path/filepath"
	"sort"
)

var cmdLink = &Command{
	Names: []string{"link"},
	Usage: []string{
		"link -l <layer> [-p <path>] <src> <dst>",
	},
	Help: `
create a link for the specified path

  link is used to create a link of <src> to <dst>, and will be managed by
  update. If <src> is not found on update, it will be ignored without error.

  The value of flag --path is a list of directories like PATH or GOPATH
  environment variables, and it is used to search <src>.

  You can refer environment variables in <path> and <src>. Supported formats
  are ${var} and $var.

options:

  -l, --layer    a layer
  -p, --path     a list of directories to search <src>
`,
}

var (
	linkL string
	linkP string
)

func init() {
	cmdLink.Flag.StringVar(&linkL, "l", "", "")
	cmdLink.Flag.StringVar(&linkL, "layer", "", "")
	cmdLink.Flag.StringVar(&linkP, "p", "", "")
	cmdLink.Flag.StringVar(&linkP, "path", "", "")

	cmdLink.Run = runLink
}

func runLink(ui UI, args []string) error {
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
	case linkL == "":
		return FlagError("flag --layer is required")
	default:
		if len(args) != 2 {
			return ErrArg
		}
		l, err := repo.LayerOf(linkL)
		switch {
		case err != nil:
			return err
		case 0 < len(l.Layers):
			return fmt.Errorf("layer '%s' is abstract", l.Path())
		}
		dst, err := wc.Rel('.', args[1])
		if err != nil {
			return err
		}
		switch typ := repo.Find(l, dst); typ {
		case "":
		case "dir", "file":
			return fmt.Errorf("'%s' already exists!", dst)
		default:
			return fmt.Errorf("%s '%s' already exists!", typ, dst)
		}
		path := filepath.SplitList(linkP)
		for i, p := range path {
			path[i] = filepath.ToSlash(filepath.Clean(p))
		}
		src := filepath.ToSlash(filepath.Clean(args[0]))
		dir, dst := splitPath(dst)
		if l.Links == nil {
			l.Links = make(map[string][]*Link)
		}
		l.Links[dir] = append(l.Links[dir], &Link{
			Path: path,
			Src:  src,
			Dst:  dst,
		})
		sort.Sort(linkByDst(l.Links[dir]))
	}
	return repo.Flush()
}
