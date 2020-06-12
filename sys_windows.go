//
// nazuna :: sys_windows.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna

const (
	_FSCTL_SET_REPARSE_POINT = 0x000900a4

	_SYMLINK_FLAG_RELATIVE = 0x00000001
)

type reparseDataBuffer struct {
	ReparseTag        uint32
	ReparseDataLength uint16
	Reserved          uint16
	ReparseBuffer     byte
}

type symbolicLinkReparseBuffer struct {
	SubstituteNameOffset uint16
	SubstituteNameLength uint16
	PrintNameOffset      uint16
	PrintNameLength      uint16
	Flags                uint32
	PathBuffer           [1]uint16
}

type mountPointReparseBuffer struct {
	SubstituteNameOffset uint16
	SubstituteNameLength uint16
	PrintNameOffset      uint16
	PrintNameLength      uint16
	PathBuffer           [1]uint16
}
