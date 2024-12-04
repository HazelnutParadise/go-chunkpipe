//go:build !windows

package chunkpipe

import (
	"syscall"
	"unsafe"
)

const (
	hugePageSize  = 2 * 1024 * 1024 // 2MB huge pages
	cacheLineSize = 64
	MAP_HUGETLB   = 0x40000 // Linux huge page flag
)

func (p *MemoryPool) numaAwareAlloc(size uintptr) unsafe.Pointer {
	// 對齊到 huge page
	alignedSize := (size + hugePageSize - 1) &^ (hugePageSize - 1)

	// 使用 mmap 分配記憶體
	addr, _, err := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,
		alignedSize,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_PRIVATE|MAP_ANON|MAP_HUGETLB,
		0, 0)

	if err != 0 {
		// 失敗時使用普通分配
		return p.normalAlloc(size)
	}

	return unsafe.Pointer(addr)
}
