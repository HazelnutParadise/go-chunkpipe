package chunkpipe

func (vc *valueCache[T]) setValueCache(index int, value *T) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	for index >= len(vc.cache) {
		vc.cache = append(vc.cache, make([]*T, 0, 1024)...)
	}

	vc.cache[index] = value
}

func (vc *valueCache[T]) getValueCache(index int) *T {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	if index < 0 || index >= len(vc.cache) {
		return nil
	}

	return vc.cache[index]
}

func (vc *valueCache[T]) dropFirstValueCache() {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if len(vc.cache) > 0 {
		vc.cache = vc.cache[1:]
	}
}

func (vc *valueCache[T]) dropValueCacheBefore(index int) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if index >= len(vc.cache) {
		return
	}

	vc.cache = vc.cache[index:]
}

func (vc *valueCache[T]) clearValueCache() {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	vc.cache = make([]*T, 0, 1024)
}

func (vc *valueCache[T]) deleteValueCache(index int) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	vc.cache[index] = nil
}
