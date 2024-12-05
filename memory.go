package chunkpipe

import (
	"sync/atomic"
	"unsafe"
)

// MemoryPool 是一個簡單的記憶體池實現
type MemoryPool struct {
	blocks [][]byte
	size   int64 // 改用 int64 以支持原子操作
}

// newMemoryPool 創建一個新的記憶體池
func newMemoryPool() *MemoryPool {
	return &MemoryPool{
		blocks: make([][]byte, 0),
		size:   0,
	}
}

// Alloc allocates a block of memory of the specified size
func (p *MemoryPool) Alloc(size uintptr) unsafe.Pointer {
	block := make([]byte, size)
	p.blocks = append(p.blocks, block)
	atomic.AddInt64(&p.size, int64(len(block)))
	return unsafe.Pointer(&block[0])
}

// Free releases all allocated memory
func (p *MemoryPool) Free() {
	p.blocks = make([][]byte, 0)
	atomic.StoreInt64(&p.size, 0)
}

// Size returns the total size of allocated memory
func (p *MemoryPool) Size() int {
	return int(atomic.LoadInt64(&p.size))
}

// blockCache 使用原子操作的實現
type blockCache struct {
	blocks []*Chunk[byte]
	size   int32 // 使用原子操作追蹤大小
}

var globalBlockCache = &blockCache{
	blocks: make([]*Chunk[byte], 0, 1024),
	size:   0,
}

func (c *blockCache) get() *Chunk[byte] {
	size := atomic.LoadInt32(&c.size)
	if size > 0 {
		if atomic.CompareAndSwapInt32(&c.size, size, size-1) {
			block := c.blocks[size-1]
			c.blocks = c.blocks[:size-1]
			return block
		}
	}
	return nil
}

func (c *blockCache) put(block *Chunk[byte]) {
	size := atomic.LoadInt32(&c.size)
	if size < 1024 {
		if atomic.CompareAndSwapInt32(&c.size, size, size+1) {
			block.next = nil
			block.prev = nil
			c.blocks = append(c.blocks, block)
		}
	}
}
