//
// nazuna :: util_windows.go
//
//   Copyright (c) 2013-2014 Akinori Hattori <hattya@gmail.com>
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

package nazuna

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

func RemoveAll(path string) error {
	// syscall.DeleteFile cannot remove read-only files
	err := filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case fi.Mode().Perm()&0200 == 0:
			mode := os.FileMode(0666)
			if fi.IsDir() {
				mode = 0777
			}
			if err := os.Chmod(path, mode); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func isLink(path string) bool {
	h, err := createFile(path, 0)
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)
	// hardlink
	var fi syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(h, &fi); err == nil && 1 < fi.NumberOfLinks {
		return true
	}
	// junction
	if fi.FileAttributes&_FILE_ATTRIBUTE_REPARSE_POINT == 0 {
		return false
	}
	b := make([]byte, _MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
	var retlen uint32
	if err := deviceIoControl(h, _FSCTL_GET_REPARSE_POINT, nil, b, &retlen, nil); err != nil {
		return false
	}
	switch rdb := (*reparseDataBuffer)(unsafe.Pointer(&b[0])); rdb.ReparseTag {
	case _IO_REPARSE_TAG_MOUNT_POINT:
	case _IO_REPARSE_TAG_SYMLINK:
	default:
		return false
	}
	return true
}

func linksTo(path, origin string) bool {
	h, err := createFile(path, 0)
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)
	// hardlink
	var fi syscall.ByHandleFileInformation
	switch err := syscall.GetFileInformationByHandle(h, &fi); {
	case err != nil:
		return false
	case 1 < fi.NumberOfLinks:
		h, err := createFile(origin, 0)
		if err != nil {
			return false
		}
		defer syscall.CloseHandle(h)

		var ofi syscall.ByHandleFileInformation
		if err := syscall.GetFileInformationByHandle(h, &ofi); err != nil || ofi.NumberOfLinks == 1 {
			return false
		}
		return fi.VolumeSerialNumber == ofi.VolumeSerialNumber && fi.FileIndexHigh == ofi.FileIndexHigh && fi.FileIndexLow == ofi.FileIndexLow
	}
	// junction
	if fi.FileAttributes&_FILE_ATTRIBUTE_REPARSE_POINT == 0 {
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
	b := make([]byte, _MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
	var retlen uint32
	if err := deviceIoControl(h, _FSCTL_GET_REPARSE_POINT, nil, b, &retlen, nil); err != nil {
		return false
	}
	switch rdb := (*reparseDataBuffer)(unsafe.Pointer(&b[0])); rdb.ReparseTag {
	case _IO_REPARSE_TAG_MOUNT_POINT:
		rb := (*mountPointReparseBuffer)(unsafe.Pointer(&rdb.ReparseBuffer[0]))
		start := rb.SubstituteNameOffset / 2
		end := start + rb.SubstituteNameLength/2
		path = syscall.UTF16ToString((*[0xffff]uint16)(unsafe.Pointer(&rb.PathBuffer[0]))[start:end])
	case _IO_REPARSE_TAG_SYMLINK:
		rb := (*symbolicLinkReparseBuffer)(unsafe.Pointer(&rdb.ReparseBuffer[0]))
		start := rb.SubstituteNameOffset / 2
		end := start + rb.SubstituteNameLength/2
		p := syscall.UTF16ToString((*[0xffff]uint16)(unsafe.Pointer(&rb.PathBuffer[0]))[start:end])
		if rb.Flags&_SYMLINK_FLAG_RELATIVE != 0 {
			path = filepath.Join(filepath.Dir(path), p)
		} else {
			path = p
		}
	default:
		return false
	}
	if strings.HasPrefix(path, `\??\`) {
		path = path[4:]
	}
	return path == origin
}

func link(src, dst string) error {
	linkError := func(err error) error {
		return &os.LinkError{
			Op:  "link",
			Old: src,
			New: dst,
			Err: err,
		}
	}

	if IsDir(src) {
		if _, err := os.Stat(dst); err == nil {
			return linkError(syscall.ERROR_ALREADY_EXISTS)
		}
		if err := os.MkdirAll(dst, 0777); err != nil {
			return err
		}
		h, err := createFile(dst, syscall.GENERIC_WRITE)
		if err != nil {
			return linkError(err)
		}
		defer syscall.CloseHandle(h)

		path, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		sn, _ := syscall.UTF16FromString(`\??\` + path)
		pn, _ := syscall.UTF16FromString(path)

		b := make([]byte, _MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
		var retlen uint32
		rdb := (*reparseDataBuffer)(unsafe.Pointer(&b[0]))
		rdb.ReparseTag = _IO_REPARSE_TAG_MOUNT_POINT
		rb := (*mountPointReparseBuffer)(unsafe.Pointer(&rdb.ReparseBuffer[0]))
		rb.SubstituteNameLength = uint16((len(sn) - 1) * 2)
		rb.PrintNameOffset = rb.SubstituteNameLength + 2
		rb.PrintNameLength = uint16((len(pn) - 1) * 2)
		copy((*[0xffff]uint16)(unsafe.Pointer(&rb.PathBuffer[0]))[:], append(sn, pn...))
		rdb.ReparseDataLength = 8 + rb.PrintNameOffset + rb.PrintNameLength + 2
		if err := deviceIoControl(h, _FSCTL_SET_REPARSE_POINT, b[:rdb.ReparseDataLength+8], nil, &retlen, nil); err != nil {
			return linkError(err)
		}
	} else {
		s, err := syscall.UTF16PtrFromString(src)
		if err != nil {
			return linkError(err)
		}
		d, err := syscall.UTF16PtrFromString(dst)
		if err != nil {
			return linkError(err)
		}
		if err := createHardLink(d, s, nil); err != nil {
			return linkError(err)
		}
	}
	return nil
}

func unlink(path string) error {
	if !isLink(path) {
		return &os.PathError{
			Op:   "unlink",
			Path: path,
			Err:  ErrNotLink,
		}
	}
	return os.Remove(path)
}

func createFile(path string, access uint32) (syscall.Handle, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	return syscall.CreateFile(p, access, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS|_FILE_FLAG_OPEN_REPARSE_POINT, 0)
}
