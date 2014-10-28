//
// nazuna :: vcs.go
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
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
)

type VCS interface {
	String() string
	Exec(...string) error

	Init(string) error
	Clone(string, string) error

	Add(...string) error
	List(...string) *exec.Cmd
	Update() error
}

type BaseVCS struct {
	Name string
	Cmd  string
	UI   UI
	Dir  string
}

func (v *BaseVCS) String() string {
	return v.Name
}

func (v *BaseVCS) Command(args ...string) *exec.Cmd {
	cmd := exec.Command(v.Cmd, args...)
	cmd.Dir = v.Dir
	return cmd
}

func (v *BaseVCS) Exec(args ...string) error {
	return v.UI.Exec(v.Command(args...))
}

func (v *BaseVCS) Init(string) error {
	return errors.New("VCS.Init not implemented")
}

func (v *BaseVCS) Clone(string, string) error {
	return errors.New("VCS.Clone not implemented")
}

func (v *BaseVCS) Add(...string) error {
	return errors.New("VCS.Add not implemented")
}

func (v *BaseVCS) List(...string) *exec.Cmd {
	return nil
}

func (v *BaseVCS) Update() error {
	return errors.New("VCS.Update not implemented")
}

type Git struct {
	BaseVCS
}

func newGit(ui UI, dir string) VCS {
	return &Git{BaseVCS{
		Name: "Git",
		Cmd:  "git",
		UI:   ui,
		Dir:  dir,
	}}
}

func (v *Git) Init(dir string) error {
	return v.Exec("init", "-q", dir)
}

func (v *Git) Clone(src, dst string) error {
	return v.Exec("clone", "--recursive", src, dst)
}

func (v *Git) Add(paths ...string) error {
	return v.Exec(append([]string{"add"}, paths...)...)
}

func (v *Git) List(paths ...string) *exec.Cmd {
	return v.Command(append([]string{"ls-files"}, paths...)...)
}

func (v *Git) Update() error {
	if err := v.Exec("pull"); err != nil {
		return err
	}
	return v.Exec("submodule", "update", "--init", "--recursive")
}

type Mercurial struct {
	BaseVCS
}

func newMercurial(ui UI, dir string) VCS {
	return &Mercurial{BaseVCS{
		Name: "Mercurial",
		Cmd:  "hg",
		UI:   ui,
		Dir:  dir,
	}}
}

func (v *Mercurial) Init(dir string) error {
	return v.Exec("init", dir)
}

func (v *Mercurial) Clone(src, dst string) error {
	return v.Exec("clone", src, dst)
}

func (v *Mercurial) Add(paths ...string) error {
	return v.Exec(append([]string{"add"}, paths...)...)
}

func (v *Mercurial) List(paths ...string) *exec.Cmd {
	return v.Command(append([]string{"status", "-madcn", "--config", "ui.slash=True"}, paths...)...)
}

func (v *Mercurial) Update() error {
	if err := v.Exec("pull"); err != nil {
		return err
	}
	return v.Exec("update")
}

type NewVCS func(UI, string) VCS

type vcsType struct {
	ctrlDir string
	new     NewVCS
}

var (
	mu sync.RWMutex

	vcses = map[string]*vcsType{
		"git": {".git", newGit},
		"hg":  {".hg", newMercurial},
	}
)

func RegisterVCS(cmd, ctrlDir string, new NewVCS) {
	mu.Lock()
	defer mu.Unlock()

	if new == nil {
		panic("NewVCS is nil")
	}
	if _, dup := vcses[cmd]; dup {
		panic(fmt.Sprintf("vcs '%s' already registered", cmd))
	}
	vcses[cmd] = &vcsType{ctrlDir, new}
}

func FindVCS(ui UI, cmd, dir string) (VCS, error) {
	mu.RLock()
	defer mu.RUnlock()

	if v, ok := vcses[cmd]; ok {
		return v.new(ui, dir), nil
	}
	return nil, fmt.Errorf("unknown vcs '%s'", cmd)
}

func VCSFor(ui UI, dir string) (VCS, error) {
	mu.RLock()
	defer mu.RUnlock()

	for _, v := range vcses {
		if IsDir(filepath.Join(dir, v.ctrlDir)) {
			vcs := v.new(ui, dir)
			return vcs, nil
		}
	}
	return nil, fmt.Errorf("unknown vcs for directory '%s'", dir)
}
