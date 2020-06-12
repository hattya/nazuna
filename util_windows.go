//
// nazuna :: util_windows.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna

import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func IsLink(path string) bool {
	h, err := createFile(path, 0)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)

	var fi windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &fi); err != nil {
		return false
	}
	// hard link
	if fi.NumberOfLinks > 1 {
		return true
	}
	// reparse point
	if fi.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT == 0 {
		return false
	}
	b := make([]byte, windows.MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
	var retlen uint32
	if err := windows.DeviceIoControl(h, windows.FSCTL_GET_REPARSE_POINT, nil, 0, &b[0], uint32(len(b)), &retlen, nil); err != nil {
		return false
	}
	switch rdb := (*reparseDataBuffer)(unsafe.Pointer(&b[0])); rdb.ReparseTag {
	case windows.IO_REPARSE_TAG_SYMLINK:
	case windows.IO_REPARSE_TAG_MOUNT_POINT:
	default:
		return false
	}
	return true
}

func LinksTo(path, origin string) bool {
	h, err := createFile(path, 0)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)

	var fi1 windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &fi1); err != nil {
		return false
	}
	// hard link
	if fi1.NumberOfLinks > 1 {
		h, err := createFile(origin, 0)
		if err != nil {
			return false
		}
		defer windows.CloseHandle(h)

		var fi2 windows.ByHandleFileInformation
		if err := windows.GetFileInformationByHandle(h, &fi2); err != nil {
			return false
		}
		return fi1.VolumeSerialNumber == fi2.VolumeSerialNumber && fi1.FileIndexHigh == fi2.FileIndexHigh && fi1.FileIndexLow == fi2.FileIndexLow
	}
	// reparse point
	if fi1.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT == 0 {
		return false
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return false
	}
	origin, err = filepath.Abs(origin)
	if err != nil {
		return false
	}
	b := make([]byte, windows.MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
	var retlen uint32
	if err := windows.DeviceIoControl(h, windows.FSCTL_GET_REPARSE_POINT, nil, 0, &b[0], uint32(len(b)), &retlen, nil); err != nil {
		return false
	}
	switch rdb := (*reparseDataBuffer)(unsafe.Pointer(&b[0])); rdb.ReparseTag {
	case windows.IO_REPARSE_TAG_SYMLINK:
		rb := (*symbolicLinkReparseBuffer)(unsafe.Pointer(&rdb.ReparseBuffer))
		pb := (*[0xffff]uint16)(unsafe.Pointer(&rb.PathBuffer[0]))
		p := windows.UTF16ToString(pb[rb.SubstituteNameOffset/2 : (rb.SubstituteNameOffset+rb.SubstituteNameLength)/2])
		if rb.Flags&_SYMLINK_FLAG_RELATIVE != 0 {
			path = filepath.Join(filepath.Dir(path), p)
		} else {
			path = p
		}
	case windows.IO_REPARSE_TAG_MOUNT_POINT:
		rb := (*mountPointReparseBuffer)(unsafe.Pointer(&rdb.ReparseBuffer))
		pb := (*[0xffff]uint16)(unsafe.Pointer(&rb.PathBuffer[0]))
		path = windows.UTF16ToString(pb[rb.SubstituteNameOffset/2 : (rb.SubstituteNameOffset+rb.SubstituteNameLength)/2])
	default:
		return false
	}
	if strings.HasPrefix(path, `\??\`) {
		path = path[4:]
		switch {
		case len(path) >= 2 && path[1] == ':':
		case len(path) >= 4 && path[:4] == `UNC\`:
			path = `\\` + path[4:]
		}
	}
	return path == origin
}

func CreateLink(src, dst string) error {
	if IsDir(src) {
		return createMountPoint(src, dst)
	}
	return os.Link(src, dst)
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

func createFile(path string, access uint32) (windows.Handle, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return windows.InvalidHandle, err
	}
	return windows.CreateFile(p, access, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
}

func createMountPoint(src, dst string) error {
	linkErr := func(err error) error {
		return &os.LinkError{
			Op:  "link",
			Old: src,
			New: dst,
			Err: err,
		}
	}

	if _, err := os.Stat(dst); err == nil {
		return linkErr(windows.ERROR_ALREADY_EXISTS)
	}
	if err := os.MkdirAll(dst, 0777); err != nil {
		return linkErr(err)
	}
	h, err := createFile(dst, windows.GENERIC_WRITE)
	if err != nil {
		return linkErr(err)
	}
	defer windows.CloseHandle(h)

	path, err := filepath.Abs(src)
	if err != nil {
		return linkErr(err)
	}
	sn, err := windows.UTF16FromString(`\??\` + path)
	if err != nil {
		return linkErr(err)
	}
	pn, err := windows.UTF16FromString(path)
	if err != nil {
		return linkErr(err)
	}

	b := make([]byte, windows.MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
	var retlen uint32
	rdb := (*reparseDataBuffer)(unsafe.Pointer(&b[0]))
	rdb.ReparseTag = windows.IO_REPARSE_TAG_MOUNT_POINT
	rb := (*mountPointReparseBuffer)(unsafe.Pointer(&rdb.ReparseBuffer))
	rb.SubstituteNameLength = uint16((len(sn) - 1) * 2)
	rb.PrintNameOffset = rb.SubstituteNameLength + 2
	rb.PrintNameLength = uint16((len(pn) - 1) * 2)
	copy((*[0xffff]uint16)(unsafe.Pointer(&rb.PathBuffer[0]))[:], append(sn, pn...))
	rdb.ReparseDataLength = 8 + rb.PrintNameOffset + rb.PrintNameLength + 2
	if err := windows.DeviceIoControl(h, _FSCTL_SET_REPARSE_POINT, &b[0], uint32(rdb.ReparseDataLength+8), nil, 0, &retlen, nil); err != nil {
		return linkErr(err)
	}
	return nil
}
