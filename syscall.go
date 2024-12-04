package chunkpipe

import (
	"unsafe"
)

// 記憶體保護常數 (通用值)
const (
	PROT_READ  = 0x1
	PROT_WRITE = 0x2
	PROT_EXEC  = 0x4

	MAP_SHARED  = 0x1
	MAP_PRIVATE = 0x2
	MAP_FIXED   = 0x10
	MAP_ANON    = 0x20
)

// 系統呼叫介面
type sysCaller interface {
	mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (uintptr, error)
	munmap(addr uintptr, length uintptr) error
}

var defaultSysCaller sysCaller

func init() {
	defaultSysCaller = &stdSysCaller{}
}

// 標準系統呼叫實作 (使用 Go 的標準記憶體分配)
type stdSysCaller struct{}

func (s *stdSysCaller) mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (uintptr, error) {
	mem := make([]byte, length)
	return uintptr(unsafe.Pointer(&mem[0])), nil
}

func (s *stdSysCaller) munmap(addr uintptr, length uintptr) error {
	return nil
}
