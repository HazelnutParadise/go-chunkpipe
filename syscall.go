package chunkpipe

// CPU 緩存行大小
const CacheLineSize = 64

// 記憶體對齊
func align(size uintptr, alignment uintptr) uintptr {
	return (size + alignment - 1) &^ (alignment - 1)
}
