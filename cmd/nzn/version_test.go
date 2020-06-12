//
// nazuna/cmd/nzn :: version_test.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"fmt"
	"testing"

	"github.com/hattya/nazuna"
)

var versionOut = fmt.Sprintf("nzn version %v", nazuna.Version)

func TestVersion(t *testing.T) {
	s := script{
		{
			cmd: []string{"nzn", "--version"},
			out: versionOut,
		},
		{
			cmd: []string{"nzn", "version"},
			out: versionOut,
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}
