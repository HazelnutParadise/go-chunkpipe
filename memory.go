package chunkpipe

import (
	"sync/atomic"
	"unsafe"
)

// MemoryPool 使用指標實現的記憶體池
type MemoryPool struct {
	blocks []*block
	size   int64
}

// block 代表一個記憶體塊
type block struct {
	data unsafe.Pointer
	size uintptr
}

// newMemoryPool 創建一個新的記憶體池
func newMemoryPool() *MemoryPool {
	return &MemoryPool{
		blocks: make([]*block, 0),
		size:   0,
	}
}

// Alloc 分配指定大小的記憶體
func (p *MemoryPool) Alloc(size uintptr) unsafe.Pointer {
	// 直接分配記憶體
	ptr := unsafe.Pointer((&make([]byte, size)[0]))

	// 創建新的記憶體塊
	b := &block{
		data: ptr,
		size: size,
	}

	// 添加到blocks並更新大小
	p.blocks = append(p.blocks, b)
	atomic.AddInt64(&p.size, int64(size))

	return ptr
}

// Free 釋放所有分配的記憶體
func (p *MemoryPool) Free() {
	// 清空blocks
	p.blocks = make([]*block, 0)
	atomic.StoreInt64(&p.size, 0)
}

// Size 返回已分配的記憶體總大小
func (p *MemoryPool) Size() int {
	return int(atomic.LoadInt64(&p.size))
}

// blockCache 使用指標實現的塊緩存
type blockCache struct {
	head *cacheNode
	tail *cacheNode
	size int32
}

// cacheNode 代表緩存中的一個節點
type cacheNode struct {
	chunk *Chunk[byte]
	next  *cacheNode
}

var globalBlockCache = &blockCache{
	head: nil,
	tail: nil,
	size: 0,
}

func (c *blockCache) get() *Chunk[byte] {
	size := atomic.LoadInt32(&c.size)
	if size > 0 {
		if atomic.CompareAndSwapInt32(&c.size, size, size-1) {
			if c.head == nil {
				return nil
			}
			chunk := c.head.chunk
			c.head = c.head.next
			if c.head == nil {
				c.tail = nil
			}
			return chunk
		}
	}
	return nil
}

func (c *blockCache) put(chunk *Chunk[byte]) {
	size := atomic.LoadInt32(&c.size)
	if size < 1024 {
		if atomic.CompareAndSwapInt32(&c.size, size, size+1) {
			chunk.next = nil
			chunk.prev = nil

			node := &cacheNode{
				chunk: chunk,
				next:  nil,
			}

			if c.tail == nil {
				c.head = node
				c.tail = node
			} else {
				c.tail.next = node
				c.tail = node
			}
		}
	}
}
