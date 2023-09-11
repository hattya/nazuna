//
// nazuna :: util_unix.go
//
//   Copyright (c) 2013-2023 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

//go:build unix

package nazuna

import (
	"os"
	"path/filepath"
)

func IsLink(path string) bool {
	fi, err := os.Lstat(path)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

func LinksTo(path, origin string) bool {
	r, err := os.Readlink(path)
	if err != nil {
		return false
	}
	return filepath.Join(filepath.Dir(path), r) == origin
}

func CreateLink(src, dst string) error {
	rel, err := filepath.Rel(filepath.Dir(dst), src)
	if err != nil {
		rel = src
	}
	return os.Symlink(rel, dst)
}

func Unlink(path string) error {
	if !IsLink(path) {
		return &os.PathError{
			Op:   "unlink",
			Path: path,
			Err:  ErrNotLink,
		}
	}
	return os.Remove(path)
}
