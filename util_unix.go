//
// nazuna :: util_unix.go
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
// +build: !plan9,!windows

package nazuna

import (
	"os"
	"path/filepath"
)

func isLink(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

func linksTo(src, dest string) bool {
	if !isLink(src) {
		return false
	}
	r, err := os.Readlink(src)
	if err != nil {
		return false
	}
	return filepath.Join(filepath.Dir(src), r) == dest
}

func link(src, dest string) error {
	rel, err := filepath.Rel(filepath.Dir(dest), src)
	if err != nil {
		rel = src
	}
	return os.Symlink(rel, dest)
}

func unlink(path string) error {
	if !isLink(path) {
		return &os.PathError{"unlink", path, errNotLink}
	}
	return os.Remove(path)
}
