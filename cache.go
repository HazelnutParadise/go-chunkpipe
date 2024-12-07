package chunkpipe

func (cp *ChunkPipe[T]) setValueCache(index int, value T) {
	cp.valueCache.Store(index, value)
}

func (cp *ChunkPipe[T]) getValueCache(index int) *T {
	cached, ok := cp.valueCache.Load(index)
	if !ok {
		return nil
	}

	typedCache := cached.(T)
	return &typedCache
}

func (cp *ChunkPipe[T]) dropFirstValueCache() {
	count := 0
	// 修正索引
	cp.valueCache.Range(func(key, value any) bool {
		cp.valueCache.Store(key.(int)-1, value)
		count++
		return true
	})
	// 移除最後一個
	cp.valueCache.Delete(count - 1)
}

func (cp *ChunkPipe[T]) dropLastValueCache() {
	count := 0
	cp.valueCache.Range(func(key, value any) bool {
		count++
		return true
	})
	cp.valueCache.Delete(count - 1)
}

func (cp *ChunkPipe[T]) clearValueCache() {
	cp.valueCache.Range(func(key, value any) bool {
		cp.valueCache.Delete(key)
		return true
	})
}

func (cp *ChunkPipe[T]) deleteValueCache(index int) {
	cp.valueCache.Delete(index)
}
