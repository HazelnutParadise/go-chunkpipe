package chunkpipe

import (
	"sync"
	"unsafe"
)

// MemoryPool 是一個簡單的記憶體池實現
type MemoryPool struct {
	mu     sync.Mutex
	blocks [][]byte
	size   int
}

// NewMemoryPool 創建一個新的記憶體池
func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		blocks: make([][]byte, 0),
		size:   0,
	}
}

// Alloc allocates a block of memory of the specified size
func (p *MemoryPool) Alloc(size uintptr) unsafe.Pointer {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Allocate a new block
	block := make([]byte, size)
	p.blocks = append(p.blocks, block)
	p.size += len(block)

	return unsafe.Pointer(&block[0])
}

// Free releases all allocated memory
func (p *MemoryPool) Free() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.blocks = make([][]byte, 0)
	p.size = 0
}

// Size returns the total size of allocated memory
func (p *MemoryPool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.size
}

func (p *MemoryPool) normalAlloc(size uintptr) unsafe.Pointer {
	block := make([]byte, size)
	p.mu.Lock()
	p.blocks = append(p.blocks, block)
	p.size += len(block)
	p.mu.Unlock()
	return unsafe.Pointer(&block[0])
}

// 增加塊緩存
type blockCache struct {
	blocks []*Chunk[byte]
	mu     sync.Mutex
}

var globalBlockCache = &blockCache{
	blocks: make([]*Chunk[byte], 0, 1024),
}

func (c *blockCache) get() *Chunk[byte] {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.blocks) > 0 {
		block := c.blocks[len(c.blocks)-1]
		c.blocks = c.blocks[:len(c.blocks)-1]
		return block
	}
	return nil
}

func (c *blockCache) put(block *Chunk[byte]) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.blocks) < 1024 {
		block.next = nil
		block.prev = nil
		c.blocks = append(c.blocks, block)
	}
}
