//go:build windows
// +build windows

package chunkpipe

import (
	"syscall"
)

// ... 複製相同的常量定義 ...

// Windows 版本的 mmap
func mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (uintptr, error) {
	handle, err := syscall.CreateFileMapping(syscall.Handle(fd), nil,
		uint32(prot),
		uint32(length>>32),
		uint32(length&0xffffffff),
		nil)
	if err != nil {
		return 0, err
	}

	addr, err = syscall.MapViewOfFile(handle,
		syscall.FILE_MAP_WRITE|syscall.FILE_MAP_READ,
		0, 0, uintptr(length))
	if err != nil {
		syscall.CloseHandle(handle)
		return 0, err
	}
	return addr, nil
}

// Windows 版本的 munmap
func munmap(addr uintptr, length uintptr) error {
	return syscall.UnmapViewOfFile(addr)
}

// ... 其他代碼保持不變 ...
