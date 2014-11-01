//
// nazuna :: remote.go
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
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrRemote = errors.New("unknown remote")

type Remote struct {
	VCS  string
	URI  string
	Root string
	Path string

	ui  UI
	src string
}

func NewRemote(ui UI, src string) (*Remote, error) {
	for _, rh := range remoteHandlers {
		if !strings.HasPrefix(src, rh.Prefix) {
			continue
		}
		m := rh.rx.FindStringSubmatch(src)
		if m == nil {
			continue
		}
		g := map[string]string{
			"vcs": rh.VCS,
		}
		for i, n := range rh.rx.SubexpNames() {
			if n != "" && g[n] == "" {
				g[n] = m[i]
			}
		}
		g["uri"] = rh.Scheme + "://" + g["root"]
		if rh.Check != nil {
			if err := rh.Check(g); err != nil {
				return nil, err
			}
		}
		r := &Remote{
			VCS:  g["vcs"],
			URI:  g["uri"],
			Root: g["root"],
			Path: g["path"],
			ui:   ui,
			src:  src,
		}
		return r, nil
	}
	return nil, ErrRemote
}

func (r *Remote) Clone(base, dst string) error {
	vcs, err := FindVCS(r.ui, r.VCS, base)
	if err != nil {
		return fmt.Errorf("cannot detect remote vcs for %v", r.src)
	}
	return vcs.Clone(r.URI, dst)
}

func (r *Remote) Update(dir string) error {
	vcs, err := VCSFor(r.ui, dir)
	if err != nil {
		return err
	}
	return vcs.Update()
}

type RemoteHandler struct {
	Prefix string
	Expr   string
	VCS    string
	Scheme string
	Check  func(map[string]string) error

	rx *regexp.Regexp
}

var remoteHandlers = []*RemoteHandler{
	{
		Prefix: "github.com/",
		Expr:   `^(?P<root>github\.com/[^/]+/[^/]+)(?P<path>.*)$`,
		VCS:    "git",
		Scheme: "https",
	},
	{
		Prefix: "bitbucket.org/",
		Expr:   `^(?P<root>bitbucket\.org/(?P<repo>[^/]+/[^/]+))(?P<path>.*)$`,
		Scheme: "https",
		Check:  bitbucket,
	},
}

func init() {
	for _, r := range remoteHandlers {
		r.rx = regexp.MustCompile(r.Expr)
	}
}

func bitbucket(m map[string]string) error {
	var resp struct {
		SCM string
	}
	uri := "https://api.bitbucket.org/1.0/repositories/" + m["repo"]
	data, err := httpGet(uri)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("%v: %v", uri, err)
	}
	m["vcs"] = resp.SCM
	if resp.SCM == "git" {
		m["uri"] += ".git"
	}
	return nil
}
