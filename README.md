# Nazuna

Nazuna is a layered dotfiles management tool.

[![pkg.go.dev](https://pkg.go.dev/badge/github.com/hattya/nazuna)](https://pkg.go.dev/github.com/hattya/nazuna)
[![GitHub Actions](https://github.com/hattya/nazuna/workflows/CI/badge.svg)](https://github.com/hattya/nazuna/actions?query=workflow:CI)
[![Semaphore](https://semaphoreci.com/api/v1/hattya/nazuna/branches/master/badge.svg)](https://semaphoreci.com/hattya/nazuna)
[![Appveyor](https://ci.appveyor.com/api/projects/status/2eg4vbro37mhsdk0/branch/master?svg=true)](https://ci.appveyor.com/project/hattya/nazuna)
[![Codecov](https://codecov.io/gh/hattya/nazuna/branch/master/graph/badge.svg)](https://codecov.io/gh/hattya/nazuna)


## Installation

```console
$ go get -u github.com/hattya/nazuna/cmd/nzn
```


## Usage

```console
$ nzn init --vcs git
$ nzn layer -c master
$ cp .gitconfig .nzn/r/master
$ nzn vcs add .
$ nzn vcs commit -m "Initial import"
[master (root-commit) 1234567] Initial import
 2 files changed, 8 insertions(+)
 create mode 100644 master/.gitconfig
 create mode 100644 nazuna.json
$ rm .gitconfig
$ nzn update
link .gitconfig --> master
1 updated, 0 removed, 0 failed
$ readlink .gitconfig
.nzn/r/master/.gitconfig
```


## License

Nazuna is distributed under the terms of the MIT License.
