//
// nazuna :: export_test.go
//
//   Copyright (c) 2014-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna

var (
	SortKeys  = sortKeys
	Marshal   = marshal
	Unmarshal = unmarshal
)

func (l *Layer) SetAbst(abst *Layer) {
	l.abst = abst
}

func (l *Layer) SetRepo(repo *Repository) {
	l.repo = repo
}

func (repo *Repository) Root() string {
	return repo.root
}
