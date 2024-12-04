package chunkpipe

import (
	"unsafe"
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

	// 使用系統呼叫介面分配記憶體
	addr, err := defaultSysCaller.mmap(0, size,
		PROT_READ|PROT_WRITE,
		MAP_PRIVATE|MAP_ANON|
			int(node)<<24, // NUMA 節點選擇
		-1, 0)

	if err != nil {
		// 失敗時使用標準分配
		mem := make([]byte, size)
		return unsafe.Pointer(&mem[0])
	}

	return unsafe.Pointer(addr)
}
