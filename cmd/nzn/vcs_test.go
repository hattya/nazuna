//
// nazuna/cmd/nzn :: vcs_test.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package main

import (
	"testing"

	"github.com/hattya/go.cli"
)

func TestVCS(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"nzn", "vcs", "--version"},
			out: cli.Dedent(`
				git version \d.+ (re)
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestVCSError(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "vcs", "--version"},
			out: cli.Dedent(`
				nzn: no repository found in '.+' \(\.nzn not found\)! (re)
				[1]
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}
