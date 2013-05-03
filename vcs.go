//
// nazuna :: vcs.go
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
	"os/exec"
	"strings"
)

type VCS struct {
	Name string
	Cmd  string

	InitCmd string
}

func (v *VCS) String() string {
	return v.Name
}

func (v *VCS) Init(path string) *exec.Cmd {
	args := v.expand(v.InitCmd, "path", path)
	return exec.Command(v.Cmd, args...)
}

func (v *VCS) expand(cmdline string, kv ...string) []string {
	m := make(map[string]string)
	for i := 0; i < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	args := strings.Fields(cmdline)
	for i, a := range args {
		if strings.Contains(a, "{") {
			for k, v := range m {
				a = strings.Replace(a, "{"+k+"}", v, -1)
			}
		}
		args[i] = a
	}
	return args
}

var VCSes = []*VCS{
	vcsGit,
	vcsHg,
}

var vcsGit = &VCS{
	Name: "Git",
	Cmd:  "git",

	InitCmd: "init -q {path}",
}

var vcsHg = &VCS{
	Name: "Mercurial",
	Cmd:  "hg",

	InitCmd: "init {path}",
}

type VCSError struct {
	Cmd  string
	List []string
}

func (e *VCSError) Error() string {
	if len(e.List) == 0 {
		return fmt.Sprintf("unknown vcs '%s'", e.Cmd)
	}
	return fmt.Sprintf("vcs '%s' is ambiguous:\n    %s", e.Cmd, strings.Join(e.List, " "))
}

func FindVCS(cmd string) (vcs *VCS, err error) {
	set := make(map[string]*VCS)
	cmd = strings.ToLower(cmd)
	for _, v := range VCSes {
		if strings.HasPrefix(strings.ToLower(v.Cmd), cmd) {
			set[v.Cmd] = v
		}
	}

	switch len(set) {
	case 0:
		err = &VCSError{Cmd: cmd}
	case 1:
		for _, vcs = range set {
		}
	default:
		list := make([]string, len(set))
		i := 0
		for n, _ := range set {
			list[i] = n
			i++
		}
		err = &VCSError{cmd, list}
	}
	return
}
