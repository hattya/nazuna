//
// nazuna/cmd/nzn :: subrepo_test.go
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
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hattya/go.cli"
	"github.com/hattya/nazuna/internal/test"
)

func TestSubrepo(t *testing.T) {
	sh, err := newShell()
	if err != nil {
		t.Fatal(err)
	}
	defer sh.exit()

	fs := http.FileServer(http.Dir(filepath.Join(sh.dir, "public")))
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/2.0/repositories/") {
			fmt.Fprintf(w, `{"owner":{"username":"%v"},"scm":"git"}`, strings.Split(r.RequestURI[18:], "/")[0])
		} else {
			fs.ServeHTTP(w, r)
		}
	}))
	defer ts.Close()

	c := http.DefaultClient
	defer func() { http.DefaultClient = c }()
	http.DefaultClient = test.NewHTTPClient(ts.Listener.Addr().String())

	sh.gitconfig["http.sslVerify"] = "false"
	sh.gitconfig["merge.stat"] = "false"
	sh.gitconfig["url."+ts.URL+"/vim-pathogen/.git.insteadOf"] = "https://github.com/tpope/vim-pathogen"
	sh.gitconfig["url."+ts.URL+"/editorconfig-vim/.git.insteadOf"] = "https://bitbucket.org/editorconfig/editorconfig-vim.git"

	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"git", "init", "-q", "$public/vim-pathogen"},
		},
		{
			cmd: []string{"cd", "$public/vim-pathogen"},
		},
		{
			cmd: []string{"mkdir", "autoload"},
		},
		{
			cmd: []string{"touch", "autoload/pathogen.vim"},
		},
		{
			cmd: []string{"git", "add", "."},
		},
		{
			cmd: []string{"git", "commit", "-qm", "."},
		},
		{
			cmd: []string{"git", "update-server-info"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"git", "init", "-q", "$public/editorconfig-vim"},
		},
		{
			cmd: []string{"cd", "$public/editorconfig-vim"},
		},
		{
			cmd: []string{"mkdir", "plugin"},
		},
		{
			cmd: []string{"touch", "plugin/editorconfig.vim"},
		},
		{
			cmd: []string{"git", "add", "."},
		},
		{
			cmd: []string{"git", "commit", "-qm", "."},
		},
		{
			cmd: []string{"git", "update-server-info"},
		},
		{
			cmd: []string{"cd", "$tempdir"},
		},
		{
			cmd: []string{"export", "GOROOT=$public/go"},
		},
		{
			cmd: []string{"mkdir", "$public/go/misc/vim"},
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "bitbucket.org/editorconfig/editorconfig-vim", ".vim/bundle/"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				0 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "link", "-l", "a", "$GOROOT/misc/vim", ".vim/bundle/golang"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-u"},
			out: cli.Dedent(`
				* bitbucket.org/editorconfig/editorconfig-vim
				Cloning into '.nzn/sub/bitbucket.org/editorconfig/editorconfig-vim'...
				* github.com/tpope/vim-pathogen
				Cloning into '.nzn/sub/github.com/tpope/vim-pathogen'...
			`),
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vim/bundle/editorconfig-vim --> bitbucket.org/editorconfig/editorconfig-vim
				link .vim/bundle/golang/ --> .+` + quote("/go/misc/vim/") + ` (re)
				link .vim/bundle/vim-pathogen --> github.com/tpope/vim-pathogen
				3 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"cd", "$public/editorconfig-vim"},
		},
		{
			cmd: []string{"mkdir", "autoload"},
		},
		{
			cmd: []string{"touch", "autoload/editorconfig.vim"},
		},
		{
			cmd: []string{"git", "add", "."},
		},
		{
			cmd: []string{"git", "commit", "-qm", "."},
		},
		{
			cmd: []string{"git", "update-server-info"},
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-u"},
			out: cli.Dedent(`
				* bitbucket.org/editorconfig/editorconfig-vim
				From https://127.0.0.1:\d+/editorconfig-vim (re)
				   [[:alnum:]]+\.\.[[:alnum:]]+  master\s+ -> origin/master (re)
				Updating [[:alnum:]]+\.\.[[:alnum:]]+ (re)
				Fast-forward
				* github.com/tpope/vim-pathogen
				Already up.to.date\. (re)
			`),
		},
	}
	if err := sh.run(s); err != nil {
		t.Error(err)
	}
}

func TestSubrepoError(t *testing.T) {
	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "subrepo"},
			out: cli.Dedent(`
				nzn: no repository found in '.+' \(\.nzn not found\)! (re)
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"touch", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "subrepo"},
			out: cli.Dedent(`
				nzn: ` + path(".nzn/state.json") + `: unexpected end of JSON input
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".nzn/state.json"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-a"},
			out: cli.Dedent(`
				nzn subrepo: --layer flag is required
				usage: nzn subrepo -l <layer> -a <repository> <path>
				   or: nzn subrepo -u

				manage subrepositories

				  subrepo is used to manage external repositories.

				  subrepo can associate <repository> to <path> by --add flag. If <path> ends
				  with a path separator, it will be associated as the basename of <repository>
				  under <path>.

				  subrepo can clone or update the repositories in the working copy by --update
				  flag.

				options:

				  -a, --add              add <repository> to <path>
				  -l, --layer <layer>    layer name
				  -u, --update           clone or update repositories

				[2]
			`),
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a"},
			out: cli.Dedent(`
				nzn: invalid arguments
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
			out: cli.Dedent(`
				nzn: layer 'a' does not exist!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/a/.vim/bundle/vim-pathogen/autoload"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.vim/bundle/vim-pathogen/autoload/pathogen.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
			out: cli.Dedent(`
				nzn: '.vim/bundle/vim-pathogen' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-qrf", "a/.vim"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", "../dst"},
			out: cli.Dedent(`
				nzn: '../dst' is not under root
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
			out: cli.Dedent(`
				nzn: subrepo '.vim/bundle/vim-pathogen' already exists!
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/kien/ctrlp.vim", ".vim/bundle/ctrlp.vim"},
		},
		{
			cmd: []string{"mkdir", ".nzn/sub/github.com/kien/ctrlp.vim"},
		},
		{
			cmd: []string{"mkdir", ".nzn/sub/github.com/tpope/vim-pathogen"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				link .vim/bundle/ctrlp.vim --> github.com/kien/ctrlp.vim
				link .vim/bundle/vim-pathogen --> github.com/tpope/vim-pathogen
				2 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "b"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "b", "-a", "github.com/kien/ctrlp.vim", ".vim/bundle/"},
		},
		{
			cmd: []string{"rm", ".vim/bundle/ctrlp.vim"},
		},
		{
			cmd: []string{"mkdir", ".nzn/sub/_"},
		},
		{
			cmd: []string{"ln", "-s", ".nzn/sub/_", ".vim/bundle/ctrlp.vim"},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				unlink .vim/bundle/ctrlp.vim -/- github.com/kien/ctrlp.vim
				nzn: not linked to 'github.com/kien/ctrlp.vim'
				[1]
			`),
		},
		{
			cmd: []string{"rm", ".vim/bundle/ctrlp.vim"},
		},
		{
			cmd: []string{"rm", ".nzn/sub/_"},
		},
		{
			cmd: []string{"mkdir", ".nzn/r/b/.vim/bundle/ctrlp.vim/plugin"},
		},
		{
			cmd: []string{"touch", ".nzn/r/b/.vim/bundle/ctrlp.vim/plugin/ctrlp.vim"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"nzn", "update"},
			out: cli.Dedent(`
				warning: subrepo: '.vim/bundle/ctrlp.vim' exists in the repository
				link .vim/bundle/ctrlp.vim/ --> b
				1 updated, 0 removed, 0 failed
			`),
		},
		{
			cmd: []string{"nzn", "layer", "-c", "c/1"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "c", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
			out: cli.Dedent(`
				nzn: layer 'c' is abstract
				[1]
			`),
		},
	}
	if err := s.exec(); err != nil {
		t.Error(err)
	}
}

func TestSubrepoUpdateError(t *testing.T) {
	sh, err := newShell()
	if err != nil {
		t.Fatal(err)
	}
	defer sh.exit()

	fs := http.FileServer(http.Dir(filepath.Join(sh.dir, "public")))
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
	defer ts.Close()

	c := http.DefaultClient
	defer func() { http.DefaultClient = c }()
	http.DefaultClient = test.NewHTTPClient(ts.Listener.Addr().String())

	sh.gitconfig["http.sslVerify"] = "false"
	sh.gitconfig["url."+ts.URL+"/vim-pathogen/.git.insteadOf"] = "https://github.com/tpope/vim-pathogen"

	s := script{
		{
			cmd: []string{"setup"},
		},
		{
			cmd: []string{"git", "init", "-q", "$public/vim-pathogen"},
		},
		{
			cmd: []string{"cd", "$public/vim-pathogen"},
		},
		{
			cmd: []string{"mkdir", "autoload"},
		},
		{
			cmd: []string{"touch", "autoload/pathogen.vim"},
		},
		{
			cmd: []string{"git", "add", "."},
		},
		{
			cmd: []string{"git", "commit", "-qm", "."},
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"nzn", "init", "--vcs", "git"},
		},
		{
			cmd: []string{"nzn", "layer", "-c", "a"},
		},
		{
			cmd: []string{"touch", ".nzn/r/a/.vimrc"},
		},
		{
			cmd: []string{"nzn", "vcs", "add", "."},
		},
		{
			cmd: []string{"rm", ".nzn/r/a/.vimrc"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-u"},
			out: cli.Dedent(`
				nzn: \w+ .+` + quote("/.nzn/r/a/.vimrc") + `: .+ (re)
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "vcs", "rm", "-qf", "a/.vimrc"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "github.com/tpope/vim-pathogen", ".vim/bundle/"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-u"},
			out: cli.Dedent(`
				* github.com/tpope/vim-pathogen
				Cloning into '.nzn/sub/github.com/tpope/vim-pathogen'...
				remote: 404 page not found
				fatal: repository 'https://127.0.0.1:\d+/vim-pathogen/.git/' not found (re)
				nzn: git: exit status \d+ (re)
				[1]
			`),
		},
		{
			cmd: []string{"cd", "$public/vim-pathogen"},
		},
		{
			cmd: []string{"git", "update-server-info"},
		},
		{
			cmd: []string{"cd", "$wc"},
		},
		{
			cmd: []string{"mkdir", ".nzn/sub/github.com/tpope/vim-pathogen/autoload"},
		},
		{
			cmd: []string{"touch", ".nzn/sub/github.com/tpope/vim-pathogen/autoload/pathogen.vim"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-u"},
			out: cli.Dedent(`
				* github.com/tpope/vim-pathogen
				nzn: unknown vcs for directory '.+` + quote("/.nzn/sub/github.com/tpope/vim-pathogen") + `' (re)
				[1]
			`),
		},
		{
			cmd: []string{"rm", "-r", ".nzn/sub"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-u"},
			out: cli.Dedent(`
				* github.com/tpope/vim-pathogen
				Cloning into '.nzn/sub/github.com/tpope/vim-pathogen'...
			`),
		},
		{
			cmd: []string{"rm", "$public/vim-pathogen/.git/info/refs"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-u"},
			out: cli.Dedent(`
				* github.com/tpope/vim-pathogen
				remote: 404 page not found
				fatal: repository 'https://127.0.0.1:\d+/vim-pathogen/.git/' not found (re)
				nzn: git: exit status \d+ (re)
				[1]
			`),
		},
		{
			cmd: []string{"nzn", "subrepo", "-l", "a", "-a", "example.com/repo", ".r"},
		},
		{
			cmd: []string{"nzn", "subrepo", "-u"},
			out: cli.Dedent(`
				* example.com/repo
				nzn: unknown remote
				[1]
			`),
		},
	}
	if err := sh.run(s); err != nil {
		t.Error(err)
	}
}
