package chunkpipe

import (
	"sync"
	"syscall"
	"unsafe"
)

const (
	pageSize        = 4096
	blockSize       = 1 << 20      // 使用 1MB 塊，方便位運算
	alignMask       = ^uintptr(15) // 16字節對齊掩碼
	smallBlockSize  = 64           // 小塊大小
	mediumBlockSize = 4096         // 中等塊大小
)

type MemoryBlock struct {
	addr uintptr
	size uintptr
	used uintptr
	next *MemoryBlock
}

type MemoryPool struct {
	blocks    *MemoryBlock
	hotBlock  *MemoryBlock // 熱塊快速路徑
	pageSize  uintptr
	blockSize uintptr
	mu        sync.Mutex
}

//go:linkname mmap syscall.mmap
func mmap(addr uintptr, length uintptr, prot int, flags int, fd int, offset int64) (uintptr, error)

//go:linkname munmap syscall.munmap
func munmap(addr uintptr, length uintptr) error

func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		pageSize:  pageSize,
		blockSize: blockSize,
	}
}

func (p *MemoryPool) Alloc(size uintptr) unsafe.Pointer {
	if size <= smallBlockSize {
		return p.allocSmall(size)
	}
	// 快速路徑：小於 4K 的分配
	if size <= 4096 {
		size = (size + 15) & alignMask

		// 檢查熱塊
		if p.hotBlock != nil && p.hotBlock.size-p.hotBlock.used >= size {
			addr := p.hotBlock.addr + p.hotBlock.used
			p.hotBlock.used += size
			return unsafe.Pointer(addr)
		}

		// 分配新的熱塊
		addr, _ := mmap(0, blockSize,
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_PRIVATE|syscall.MAP_ANON,
			-1, 0)

		p.hotBlock = &MemoryBlock{
			addr: addr,
			size: blockSize,
			used: size,
		}
		return unsafe.Pointer(addr)
	}

	// 大塊直接分配
	size = (size + 15) & alignMask
	addr, _ := mmap(0, size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_PRIVATE|syscall.MAP_ANON,
		-1, 0)
	return unsafe.Pointer(addr)
}

func (p *MemoryPool) getFreeBlock(size uintptr) *MemoryBlock {
	for block := p.blocks; block != nil; block = block.next {
		if block.size-block.used >= size {
			block.used += size
			return block
		}
	}
	return nil
}

func (p *MemoryPool) Free(ptr unsafe.Pointer, size uintptr) {
	p.mu.Lock()
	defer p.mu.Unlock()

	addr := uintptr(ptr)
	size = (size + p.pageSize - 1) & ^(p.pageSize - 1)

	// 找到對應的塊
	block := p.blocks
	var prev *MemoryBlock
	for block != nil {
		if addr >= block.addr && addr < block.addr+block.size {
			// 如果整個塊都空了，就釋放它
			if block.used == block.size {
				if prev == nil {
					p.blocks = block.next
				} else {
					prev.next = block.next
				}
				munmap(block.addr, block.size)
			}
			return
		}
		prev = block
		block = block.next
	}
}

// 小塊記憶體分配
func (p *MemoryPool) allocSmall(size uintptr) unsafe.Pointer {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 對齊到 16 字節
	size = (size + 15) & alignMask

	// 檢查熱塊
	if p.hotBlock != nil && p.hotBlock.size-p.hotBlock.used >= size {
		addr := p.hotBlock.addr + p.hotBlock.used
		p.hotBlock.used += size
		return unsafe.Pointer(addr)
	}

	// 分配新的小塊
	addr, _ := mmap(0, smallBlockSize*16, // 一次分配多個小塊
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_PRIVATE|syscall.MAP_ANON,
		-1, 0)

	p.hotBlock = &MemoryBlock{
		addr: addr,
		size: smallBlockSize * 16,
		used: size,
	}
	return unsafe.Pointer(addr)
}
