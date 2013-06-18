//
// nazuna :: link.go
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
	"path/filepath"
	"sort"
)

var cmdLink = &Command{
	Names: []string{"link"},
	Usage: "link -l <layer> [-p <path>] <src> <dst>",
	Help: `
create a link for the specified path

  link is used to create a link of <src> to <dst>, and will be managed by
  update. If <src> is not found on update, it will ignore without error.

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
	linkLayer string
	linkPath  string
)

func init() {
	cmdLink.Run = runLink
	cmdLink.Flag.StringVar(&linkLayer, "l", "", "")
	cmdLink.Flag.StringVar(&linkLayer, "layer", "", "")
	cmdLink.Flag.StringVar(&linkPath, "p", "", "")
	cmdLink.Flag.StringVar(&linkPath, "path", "", "")
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
	case linkLayer == "":
		return FlagError("flag --layer is required")
	default:
		if len(args) != 2 {
			return errArg
		}
		l, err := repo.LayerOf(linkLayer)
		if err != nil {
			return err
		}
		dst, err := wc.Rel(args[1])
		if err != nil {
			return err
		}
		switch typ := repo.Find(l, dst); typ {
		case "dir", "file":
			return fmt.Errorf("'%s' already exists!", dst)
		case "link":
			return fmt.Errorf("%s '%s' already exists!", typ, dst)
		}
		path := filepath.SplitList(linkPath)
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
