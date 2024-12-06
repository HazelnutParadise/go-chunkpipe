package chunkpipe

import (
	"encoding/binary"
	"unsafe"

	"github.com/VictoriaMetrics/fastcache"
)

// memoryPool 使用 fastcache 的記憶體池
type memoryPool struct {
	cache *fastcache.Cache
}

// newMemoryPool 創建一個新的記憶體池
func newMemoryPool() *memoryPool {
	return &memoryPool{
		cache: fastcache.New(512 * 1024 * 1024), // 512MB 快取
	}
}

// Alloc 分配指定大小的記憶體
func (p *memoryPool) Alloc(size uintptr) unsafe.Pointer {
	// 使用大小作為 key
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, uint64(size))

	// 嘗試從快取獲取
	value := p.cache.Get(nil, key)
	if len(value) > 0 {
		return *(*unsafe.Pointer)(unsafe.Pointer(&value[0]))
	}

	// 分配新記憶體
	ptr := unsafe.Pointer(&make([]byte, size)[0])
	return ptr
}

// Free 釋放記憶體到快取
func (p *memoryPool) Free(ptr unsafe.Pointer, size uintptr) {
	// 準備 value
	value := make([]byte, unsafe.Sizeof(uintptr(0)))
	*(*unsafe.Pointer)(unsafe.Pointer(&value[0])) = ptr

	// 使用大小作為 key
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, uint64(size))

	// 存入快取
	p.cache.Set(key, value)
}

// Size 返回已分配的記憶體總大小
func (p *memoryPool) Size() int {
	// 使用 UpdateStats 並返回實際使用的記憶體大小
	var stats fastcache.Stats
	p.cache.UpdateStats(&stats)
	return int(stats.BytesSize)
}

// Reset 重置記憶體池
func (p *memoryPool) Reset() {
	// 重新創建快取
	p.cache = fastcache.New(512 * 1024 * 1024)
}
