//
// nazuna/cmd/nzn :: layer_test.go
//
//   Copyright (c) 2013-2018 Akinori Hattori <hattya@gmail.com>
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

package main

import (
	"testing"

	"github.com/hattya/go.cli"
)

func TestLayer(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"cd", "w"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"nzn", "layer"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"nzn", "layer"},
			out: cli.Dedent(`
				a
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b"},
		},
		{
			cmd: []string{"nzn", "layer"},
			out: cli.Dedent(`
				b
				a
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/2"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/1"},
		},
		{
			cmd: []string{"nzn", "layer"},
			out: cli.Dedent(`
				c
				    1
				    2
				b
				a
			`),
		},
		{
			cmd: []string{"nzn", "layer", "c/1"},
		},
		{
			cmd: []string{"nzn", "layer"},
			out: cli.Dedent(`
				c
				    1*
				    2
				b
				a
			`),
		},
		{
			cmd: []string{"nzn", "layer", "c/2"},
		},
		{
			cmd: []string{"nzn", "layer"},
			out: cli.Dedent(`
				c
				    1
				    2*
				b
				a
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestLayerError(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"nzn", "layer"},
			out: cli.Dedent(`
				nzn: no repository found in '.*' \(\.nzn not found\)! (re)
				[1]
			`),
		},
		{
			cmd: []string{"cd", "w"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"nzn", "layer", "-c"},
			out: cli.Dedent(`
				nzn: invalid arguments
				[1]
			`),
		},
		{
			cmd: []string{"touch", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "layer"},
			out: cli.Dedent(`
				nzn: \.nzn[/\\]state.json: unexpected end of JSON input (re)
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
			out: cli.Dedent(`
				nzn: layer 'a' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a/1"},
			out: cli.Dedent(`
				nzn: layer 'a' is not abstract
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "/"},
			out: cli.Dedent(`
				nzn: invalid layer '/'
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b/"},
			out: cli.Dedent(`
				nzn: invalid layer 'b/'
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "/1"},
			out: cli.Dedent(`
				nzn: invalid layer '/1'
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b/1"},
		},
		{
			cmd: []string{"nzn", "layer", "_", "_"},
			out: cli.Dedent(`
				nzn: invalid arguments
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "_"},
			out: cli.Dedent(`
				nzn: layer '_' does not exist!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "b"},
			out: cli.Dedent(`
				nzn: layer 'b' is abstract
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "a"},
			out: cli.Dedent(`
				nzn: layer 'a' is not abstract
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "b/1"},
		},
		{
			cmd: []string{"nzn", "layer", "b/1"},
			out: cli.Dedent(`
				nzn: layer 'b' is already '1'
				[1]
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}
