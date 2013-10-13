//
// nazuna :: subrepo.go
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

var cmdSubrepo = &Command{
	Names: []string{"subrepo"},
	Usage: []string{
		"subrepo -l <layer> -a <repository> <path>",
		"subrepo -u",
	},
	Help: `
manage subrepositories

  subrepo is used to manage external repositories.

  subrepo can associate <repository> to <path> by flag --add. If <path> ends
  with a path separator, it will be associated as the basename of <repository>
  under <path>.

  subrepo can clone or update the repositories in the working copy by flag
  --update.

options:

  -l, --layer     a layer
  -a, --add       add <repository> to <path>
  -u, --update    clone or update repositories
`,
}

var (
	subrepoL string
	subrepoA bool
	subrepoU bool
)

func init() {
	cmdSubrepo.Flag.StringVar(&subrepoL, "l", "", "")
	cmdSubrepo.Flag.StringVar(&subrepoL, "layer", "", "")
	cmdSubrepo.Flag.BoolVar(&subrepoA, "a", false, "")
	cmdSubrepo.Flag.BoolVar(&subrepoA, "add", false, "")
	cmdSubrepo.Flag.BoolVar(&subrepoU, "u", false, "")
	cmdSubrepo.Flag.BoolVar(&subrepoU, "update", false, "")

	cmdSubrepo.Run = runSubrepo
}

func runSubrepo(ui UI, args []string) error {
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
	case subrepoA:
		switch {
		case subrepoL == "":
			return FlagError("flag --layer is required")
		case len(args) != 2:
			return ErrArg
		}
		l, err := repo.LayerOf(subrepoL)
		switch {
		case err != nil:
			return err
		case 0 < len(l.Layers):
			return fmt.Errorf("layer '%s' is abstract", l.Path())
		}
		src := args[0]
		path, err := wc.Rel(args[1])
		if err != nil {
			return err
		}
		path = filepath.ToSlash(path)
		var name, dst string
		if 0 < len(args[1]) && os.IsPathSeparator(args[1][len(args[1])-1]) {
			dst = path + "/" + filepath.Base(src)
		} else {
			path, name = splitPath(path)
			dst = path + "/" + name
			if name == filepath.Base(src) {
				name = ""
			}
		}
		switch typ := repo.Find(l, dst); typ {
		case "":
		case "dir", "file":
			return fmt.Errorf("'%s' already exists!", dst)
		default:
			return fmt.Errorf("%s '%s' already exists!", typ, dst)
		}
		if l.Subrepos == nil {
			l.Subrepos = make(map[string][]*Subrepo)
		}
		l.Subrepos[path] = append(l.Subrepos[path], &Subrepo{
			Src:  src,
			Name: name,
		})
		sort.Sort(subrepoBySrc(l.Subrepos[path]))
		return repo.Flush()
	case subrepoU:
		_, err := wc.MergeLayers()
		if err != nil {
			return err
		}
		for _, e := range wc.State.WC {
			if e.Type != "subrepo" {
				continue
			}
			ui.Printf("* %s\n", e.Origin)
			r, err := NewRemote(e.Origin)
			if err != nil {
				return err
			}
			dst := repo.SubrepoFor(e.Origin[:len(e.Origin)-len(r.Path)])
			if isEmptyDir(dst) {
				dst, _ = wc.Rel(dst)
				c := r.VCS.Clone(r.URI, dst)
				c.Dir = wc.PathFor(".")
				if err := ui.Exec(c); err != nil {
					return err
				}
			} else {
				vcs, err := VCSFor(dst)
				if err != nil {
					return err
				}
				for _, c := range vcs.Update() {
					c.Dir = dst
					if err := ui.Exec(c); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
