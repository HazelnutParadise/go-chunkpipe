package chunkpipe

import (
	"unsafe"
)

const (
	PROT_READ   = 0x1
	PROT_WRITE  = 0x2
	MAP_PRIVATE = 0x2
	MAP_ANON    = 0x20
)

type stdSysCaller struct{}

func (s *stdSysCaller) mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (uintptr, error) {
	mem := make([]byte, length)
	return uintptr(unsafe.Pointer(&mem[0])), nil
}

//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, size uintptr)

//go:noescape
//go:linkname memclrNoHeapPointers runtime.memclrNoHeapPointers
func memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)

//go:linkname getcpu runtime.getcpu
func getcpu() int32

//go:noescape
//go:linkname prefetcht0 runtime.prefetcht0
func prefetcht0(addr uintptr)
