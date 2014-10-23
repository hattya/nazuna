//
// nazuna :: sys_windows.go
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
	"syscall"
	"unsafe"
)

const (
	_FILE_ATTRIBUTE_REPARSE_POINT = 0x00000400
	_FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000

	_IO_REPARSE_TAG_MOUNT_POINT = 0xa0000003
	_IO_REPARSE_TAG_SYMLINK     = 0xa000000c

	_FSCTL_SET_REPARSE_POINT = 0x000900a4
	_FSCTL_GET_REPARSE_POINT = 0x000900a8

	_SYMLINK_FLAG_RELATIVE = 0x00000001

	_MAXIMUM_REPARSE_DATA_BUFFER_SIZE = 16 * 1024
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	pCreateHardLinkW = kernel32.NewProc("CreateHardLinkW")
	pDeviceIoControl = kernel32.NewProc("DeviceIoControl")
)

type reparseDataBuffer struct {
	ReparseTag        uint32
	ReparseDataLength uint16
	Reserved          uint16
	ReparseBuffer     [14]byte
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

func createHardLink(link, path *uint16, sa *syscall.SecurityAttributes) (err error) {
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

func deviceIoControl(h syscall.Handle, iocc uint32, inbuf []byte, outbuf []byte, retlen *uint32, overlapped *syscall.Overlapped) (err error) {
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
