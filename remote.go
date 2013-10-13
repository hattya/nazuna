//
// nazuna :: remote.go
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
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrRemote = errors.New("unknown remote")

type Remote struct {
	VCS  *VCS
	URI  string
	Path string
}

func NewRemote(src string) (*Remote, error) {
	for _, rh := range RemoteHandlers {
		if !strings.HasPrefix(src, rh.Prefix) {
			continue
		}
		m := rh.re.FindStringSubmatch(src)
		if m == nil {
			continue
		}
		g := map[string]string{
			"vcs": rh.VCS,
		}
		for i, n := range rh.re.SubexpNames() {
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
		vcs, err := FindVCS(g["vcs"])
		if err != nil {
			return nil, fmt.Errorf("cannot detect remote vcs for %s", src)
		}
		r := &Remote{
			VCS:  vcs,
			URI:  g["uri"],
			Path: g["path"],
		}
		return r, nil
	}
	return nil, ErrRemote
}

type RemoteHandler struct {
	Prefix string
	Expr   string
	VCS    string
	Scheme string
	Check  func(match map[string]string) error

	re *regexp.Regexp
}

var RemoteHandlers = []*RemoteHandler{
	{
		Prefix: "github.com",
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
	for _, r := range RemoteHandlers {
		r.re = regexp.MustCompile(r.Expr)
	}
}

func bitbucket(match map[string]string) error {
	var resp struct {
		SCM string
	}
	uri := format("https://api.bitbucket.org/1.0/repositories/{repo}", match)
	data, err := httpGet(uri)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("%s: %v", uri, err)
	}
	match["vcs"] = resp.SCM
	if resp.SCM == "git" {
		match["uri"] += ".git"
	}
	return nil
}
