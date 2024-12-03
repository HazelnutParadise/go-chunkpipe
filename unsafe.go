package chunkpipe

import (
	"syscall"
	"unsafe"
	_ "unsafe"
)

//go:linkname memmove runtime.memmove
func memmove(to, from unsafe.Pointer, size uintptr)

//go:noescape
//go:linkname memclrNoHeapPointers runtime.memclrNoHeapPointers
func memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)

//go:linkname getcpu runtime.getcpu
func getcpu() int32

// NUMA 感知的記憶體分配
func (p *MemoryPool) numaAlloc(size uintptr) unsafe.Pointer {
	cpu := getcpu()
	node := cpu >> 3 // 假設每個 NUMA 節點有 8 個核心

	// 在特定 NUMA 節點上分配記憶體
	addr, _ := mmap(0, size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_PRIVATE|syscall.MAP_ANON|
			int(node)<<24, // NUMA 節點選擇
		-1, 0)
	return unsafe.Pointer(addr)
}
