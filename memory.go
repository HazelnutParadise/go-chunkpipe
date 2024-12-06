package chunkpipe

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// MemoryPool 使用分層和緩存的記憶體池
type MemoryPool struct {
	sync.Mutex
	layers    []*layer           // 不同大小的記憶體層
	freeList  [][]unsafe.Pointer // 空閒記憶體列表
	allocated int64              // 已分配的總大小
}

// layer 代表一個特定大小的記憶體層
type layer struct {
	sync.Mutex
	blockSize uintptr
	blocks    []unsafe.Pointer
	used      int32
}

// 在文件頂部的 import 後面添加
type blockCache struct {
	sync.Mutex
	head *cacheNode
	tail *cacheNode
	size int32
}

type cacheNode struct {
	chunk *Chunk[byte]
	next  *cacheNode
}

var globalBlockCache = &blockCache{}

func (c *blockCache) get() *Chunk[byte] {
	c.Lock()
	defer c.Unlock()
	if c.head == nil {
		return nil
	}
	chunk := c.head.chunk
	c.head = c.head.next
	if c.head == nil {
		c.tail = nil
	}
	atomic.AddInt32(&c.size, -1)
	return chunk
}

func (c *blockCache) put(chunk *Chunk[byte]) {
	c.Lock()
	defer c.Unlock()
	if atomic.LoadInt32(&c.size) >= 1024 {
		return
	}
	chunk.next = nil
	chunk.prev = nil
	node := &cacheNode{chunk: chunk}
	if c.tail == nil {
		c.head = node
		c.tail = node
	} else {
		c.tail.next = node
		c.tail = node
	}
	atomic.AddInt32(&c.size, 1)
}

// newMemoryPool 創建一個新的記憶體池
func newMemoryPool() *MemoryPool {
	p := &MemoryPool{
		layers:    make([]*layer, 6), // 支援 6 種不同大小
		freeList:  make([][]unsafe.Pointer, 6),
		allocated: 0,
	}

	// 初始化不同大小的層
	sizes := []uintptr{64, 256, 1024, 4096, 16384, 65536}
	for i, size := range sizes {
		p.layers[i] = &layer{
			blockSize: size,
			blocks:    make([]unsafe.Pointer, 0, 64),
		}
		p.freeList[i] = make([]unsafe.Pointer, 0, 64)
	}

	return p
}

// findLayer 找到適合大小的層
func (p *MemoryPool) findLayer(size uintptr) int {
	for i, l := range p.layers {
		if size <= l.blockSize {
			return i
		}
	}
	return -1
}

// Alloc 分配指定大小的記憶體
func (p *MemoryPool) Alloc(size uintptr) unsafe.Pointer {
	layerIndex := p.findLayer(size)
	if layerIndex == -1 {
		// 大小超過最大層，直接分配
		ptr := unsafe.Pointer((&make([]byte, size)[0]))
		atomic.AddInt64(&p.allocated, int64(size))
		return ptr
	}

	layer := p.layers[layerIndex]
	layer.Lock()
	defer layer.Unlock()

	// 先檢查空閒列表
	if len(p.freeList[layerIndex]) > 0 {
		ptr := p.freeList[layerIndex][len(p.freeList[layerIndex])-1]
		p.freeList[layerIndex] = p.freeList[layerIndex][:len(p.freeList[layerIndex])-1]
		atomic.AddInt64(&p.allocated, int64(layer.blockSize))
		return ptr
	}

	// 分配新的記憶體塊
	ptr := unsafe.Pointer((&make([]byte, layer.blockSize)[0]))
	atomic.AddInt32(&layer.used, 1)
	atomic.AddInt64(&p.allocated, int64(layer.blockSize))
	return ptr
}

// Free 釋放記憶體到空閒列表
func (p *MemoryPool) Free(ptr unsafe.Pointer, size uintptr) {
	layerIndex := p.findLayer(size)
	if layerIndex == -1 {
		atomic.AddInt64(&p.allocated, -int64(size))
		return
	}

	layer := p.layers[layerIndex]
	layer.Lock()
	defer layer.Unlock()

	// 將記憶體塊加入空閒列表
	if len(p.freeList[layerIndex]) < cap(p.freeList[layerIndex]) {
		p.freeList[layerIndex] = append(p.freeList[layerIndex], ptr)
		atomic.AddInt64(&p.allocated, -int64(layer.blockSize))
	}
}

// Size 返回已分配的記憶體總大小
func (p *MemoryPool) Size() int {
	return int(atomic.LoadInt64(&p.allocated))
}

// Reset 重置記憶體池
func (p *MemoryPool) Reset() {
	p.Lock()
	defer p.Unlock()

	for i := range p.layers {
		p.layers[i].Lock()
		p.layers[i].blocks = make([]unsafe.Pointer, 0, 64)
		p.layers[i].used = 0
		p.layers[i].Unlock()

		p.freeList[i] = make([]unsafe.Pointer, 0, 64)
	}
	atomic.StoreInt64(&p.allocated, 0)
}
