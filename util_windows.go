//
// nazuna :: util_windows.go
//
//   Copyright (c) 2013 Akinori Hattori <hattya@gmail.com>
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

const (
	FILE_ATTRIBUTE_REPARSE_POINT = 0x00000400
	FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000

	IO_REPARSE_TAG_MOUNT_POINT = 0xa0000003
	IO_REPARSE_TAG_SYMLINK     = 0xa000000c

	FSCTL_SET_REPARSE_POINT = 0x000900a4
	FSCTL_GET_REPARSE_POINT = 0x000900a8

	SYMLINK_FLAG_RELATIVE = 0x00000001

	MAXIMUM_REPARSE_DATA_BUFFER_SIZE = 16 * 1024
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pCreateHardLinkW = kernel32.NewProc("CreateHardLinkW")
	pDeviceIoControl = kernel32.NewProc("DeviceIoControl")
)

type ReparseDataBuffer struct {
	ReparseTag        uint32
	ReparseDataLength uint16
	Reserved          uint16
	ReparseBuffer     [14]byte
}

type SymbolicLinkReparseBuffer struct {
	SubstituteNameOffset uint16
	SubstituteNameLength uint16
	PrintNameOffset      uint16
	PrintNameLength      uint16
	Flags                uint32
	PathBuffer           [1]uint16
}

type MountPointReparseBuffer struct {
	SubstituteNameOffset uint16
	SubstituteNameLength uint16
	PrintNameOffset      uint16
	PrintNameLength      uint16
	PathBuffer           [1]uint16
}

func CreateHardLink(link, path *uint16, sa *syscall.SecurityAttributes) (err error) {
	r1, _, e1 := pCreateHardLinkW.Call(uintptr(unsafe.Pointer(link)), uintptr(unsafe.Pointer(path)), uintptr(unsafe.Pointer(sa)))
	if r1 == 0 {
		if e1.(syscall.Errno) != 0 {
			err = e1
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func DeviceIoControl(h syscall.Handle, iocc uint32, inbuf []byte, outbuf []byte, retlen *uint32, overlapped *syscall.Overlapped) (err error) {
	var inp, outp *byte
	if 0 < len(inbuf) {
		inp = &inbuf[0]
	}
	if 0 < len(outbuf) {
		outp = &outbuf[0]
	}
	r1, _, e1 := pDeviceIoControl.Call(uintptr(h), uintptr(iocc), uintptr(unsafe.Pointer(inp)), uintptr(len(inbuf)), uintptr(unsafe.Pointer(outp)), uintptr(len(outbuf)), uintptr(unsafe.Pointer(retlen)), uintptr(unsafe.Pointer(overlapped)))
	if r1 == 0 {
		if e1.(syscall.Errno) != 0 {
			err = e1
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func createFile(path string, access uint32) (syscall.Handle, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, err
	}
	return syscall.CreateFile(p, access, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS|FILE_FLAG_OPEN_REPARSE_POINT, 0)
}

func isLink(path string) bool {
	h, err := createFile(path, 0)
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)

	var fi syscall.ByHandleFileInformation
	if err = syscall.GetFileInformationByHandle(h, &fi); err == nil && 1 < fi.NumberOfLinks {
		return true
	}

	if fi.FileAttributes&FILE_ATTRIBUTE_REPARSE_POINT == 0 {
		return false
	}
	buf := make([]byte, MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
	var retlen uint32
	if err := DeviceIoControl(h, FSCTL_GET_REPARSE_POINT, nil, buf, &retlen, nil); err != nil {
		return false
	}
	switch rdb := (*ReparseDataBuffer)(unsafe.Pointer(&buf[0])); rdb.ReparseTag {
	case IO_REPARSE_TAG_MOUNT_POINT:
	case IO_REPARSE_TAG_SYMLINK:
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

	var fi syscall.ByHandleFileInformation
	switch err = syscall.GetFileInformationByHandle(h, &fi); {
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

	if fi.FileAttributes&FILE_ATTRIBUTE_REPARSE_POINT == 0 {
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
	buf := make([]byte, MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
	var retlen uint32
	if err := DeviceIoControl(h, FSCTL_GET_REPARSE_POINT, nil, buf, &retlen, nil); err != nil {
		return false
	}
	switch rdb := (*ReparseDataBuffer)(unsafe.Pointer(&buf[0])); rdb.ReparseTag {
	case IO_REPARSE_TAG_MOUNT_POINT:
		rb := (*MountPointReparseBuffer)(unsafe.Pointer(&rdb.ReparseBuffer[0]))
		start := rb.SubstituteNameOffset / 2
		end := start + rb.SubstituteNameLength/2
		path = syscall.UTF16ToString((*[0xffff]uint16)(unsafe.Pointer(&rb.PathBuffer[0]))[start:end])
	case IO_REPARSE_TAG_SYMLINK:
		rb := (*SymbolicLinkReparseBuffer)(unsafe.Pointer(&rdb.ReparseBuffer[0]))
		start := rb.SubstituteNameOffset / 2
		end := start + rb.SubstituteNameLength/2
		p := syscall.UTF16ToString((*[0xffff]uint16)(unsafe.Pointer(&rb.PathBuffer[0]))[start:end])
		if rb.Flags&SYMLINK_FLAG_RELATIVE != 0 {
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
	if isDir(src) {
		if isDir(dst) {
			return &os.LinkError{"link", src, dst, syscall.ERROR_ALREADY_EXISTS}
		}
		if err := os.MkdirAll(dst, 0777); err != nil {
			return err
		}
		h, err := createFile(dst, syscall.GENERIC_WRITE)
		if err != nil {
			return &os.LinkError{"link", src, dst, err}
		}
		defer syscall.CloseHandle(h)

		path, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		path = `\??\` + path
		p, err := syscall.UTF16FromString(path)
		if err != nil {
			return &os.LinkError{"link", src, dst, err}
		}
		p = append(p, 0)

		buf := make([]byte, MAXIMUM_REPARSE_DATA_BUFFER_SIZE)
		var retlen uint32
		rdb := (*ReparseDataBuffer)(unsafe.Pointer(&buf[0]))
		rdb.ReparseTag = IO_REPARSE_TAG_MOUNT_POINT
		rdb.Reserved = 0
		rb := (*MountPointReparseBuffer)(unsafe.Pointer(&rdb.ReparseBuffer[0]))
		rb.SubstituteNameOffset = 0
		rb.SubstituteNameLength = uint16((len(p) - 2) * 2)
		rb.PrintNameOffset = rb.SubstituteNameLength + 2
		rb.PrintNameLength = 0
		copy((*[0xffff]uint16)(unsafe.Pointer(&rb.PathBuffer[0]))[:], p)
		rdb.ReparseDataLength = 8 + rb.PrintNameOffset + rb.PrintNameLength + 2
		if err := DeviceIoControl(h, FSCTL_SET_REPARSE_POINT, buf[:rdb.ReparseDataLength+8], nil, &retlen, nil); err != nil {
			return &os.LinkError{"link", src, dst, err}
		}
	} else {
		s, err := syscall.UTF16PtrFromString(src)
		if err != nil {
			return &os.LinkError{"link", src, dst, err}
		}
		d, err := syscall.UTF16PtrFromString(dst)
		if err != nil {
			return &os.LinkError{"link", src, dst, err}
		}
		if err := CreateHardLink(d, s, nil); err != nil {
			return &os.LinkError{"link", src, dst, err}
		}
	}
	return nil
}

func unlink(path string) error {
	if !isLink(path) {
		return &os.PathError{"unlink", path, errNotLink}
	}
	return os.Remove(path)
}

func RemoveAll(path string) error {
	// syscall.DeleteFile cannot remove read-only files
	err := filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case fi.IsDir():
			if err := os.Chmod(path, 0777); err != nil {
				return err
			}
		default:
			if err := os.Chmod(path, 0666); err != nil {
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
