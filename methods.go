package chunkpipe

// 插入數據到 ChunkPipe，支援泛型和鏈式呼叫
func (cl *ChunkPipe[T]) Push(data []T) *ChunkPipe[T] {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if len(data) == 0 {
		return cl
	}

	off := cl.offset
	if len(cl.list) != 0 {
		off = cl.list[len(cl.list)-1].off
	}

	cl.list = append(cl.list, offset[T]{
		val: data,
		off: off + len(data),
	})

	return cl
}

func (cl *ChunkPipe[T]) Get(index int) (T, bool) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	var zero T
	if len(cl.list) == 0 || index < 0 {
		return zero, false
	}

	target := index + cl.offset
	l := 0
	r := len(cl.list) - 1

	if target >= cl.list[r].off {
		return zero, false
	}

	if cl.list[l].off > target {
		off := cl.list[0]
		val := off.val
		target = len(val) - (off.off - target)
		return val[target], true
	}

	for r-l > 1 {
		m := (r + l) >> 1
		if cl.list[m].off > target {
			r = m
		} else {
			l = m
		}
	}

	off := cl.list[r]
	val := off.val
	target = len(val) - (off.off - target)
	return val[target], true
}

// 從頭部彈出數據
func (cl *ChunkPipe[T]) PopChunkFront() ([]T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if len(cl.list) > 0 {
		cl.offset = cl.list[0].off
		ret := cl.list[0].val
		cl.list = cl.list[1:]
		return ret, true
	}
	return nil, false
}

// 從尾部彈出數據
func (cl *ChunkPipe[T]) PopChunkEnd() ([]T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if len(cl.list) > 0 {
		ret := cl.list[len(cl.list)-1].val
		cl.list = cl.list[:len(cl.list)-1]
		return ret, true
	}
	return nil, false
}

func (cl *ChunkPipe[T]) PopFront() (T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if len(cl.list) > 0 {
		val := cl.list[0].val
		ret := val[0]
		val = val[1:]
		cl.list[0].val = val
		cl.offset++
		if len(val) == 0 {
			cl.list = cl.list[1:]
		}
		return ret, true
	}
	var ret T
	return ret, false
}

// 從尾部彈出數據
func (cl *ChunkPipe[T]) PopEnd() (T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if len(cl.list) > 0 {
		val := cl.list[len(cl.list)-1].val
		ret := val[len(val)-1]
		val = val[:len(val)-1]
		cl.list[len(cl.list)-1].val = val
		cl.list[len(cl.list)-1].off--

		if len(val) == 0 {
			// remove the element
			cl.list = cl.list[:len(cl.list)-1]
		}
		return ret, true
	}
	var ret T
	return ret, false
}

// ValueSlice 返回所有值的切片
func (cl *ChunkPipe[T]) ValueSlice() []T {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	if len(cl.list) == 0 {
		return []T{}
	}

	size := cl.list[len(cl.list)-1].off - cl.offset
	// 從 cl.valuePool 中獲取切片
	slicePtr := cl.valuePool.Get().(*[]T)
	ret := *slicePtr
	// 確保切片容量足夠
	if cap(ret) < size {
		cl.valuePool.Put(slicePtr)
		ret = make([]T, size)
	} else {
		ret = ret[:size]
	}

	k := 0
	for i := range cl.list {
		for _, v := range cl.list[i].val {
			ret[k] = v
			k++
		}
	}
	return ret
}

// ChunkSlice 返回所有數據塊的切片
func (cl *ChunkPipe[T]) ChunkSlice() [][]T {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	if len(cl.list) == 0 {
		return [][]T{}
	}

	// 從 cl.chunkPool 中獲取切片
	slicePtr := cl.chunkPool.Get().(*[][]T)
	ret := *slicePtr
	// 確保切片容量足夠
	if cap(ret) < len(cl.list) {
		cl.chunkPool.Put(slicePtr)
		ret = make([][]T, len(cl.list))
	} else {
		ret = ret[:len(cl.list)]
	}

	for i := range ret {
		ret[i] = cl.list[i].val
	}
	return ret
}

func (cl *ChunkPipe[T]) size() int {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	if len(cl.list) == 0 {
		return 0
	}
	return cl.list[len(cl.list)-1].off - cl.offset
}

// ValueIter 返回值迭代器
func (cl *ChunkPipe[T]) ValueIter() *ValueIterator[T] {
	return &ValueIterator[T]{
		pos:  -1,
		pipe: cl,
	}
}

// ChunkIter 返回塊迭代器
func (cl *ChunkPipe[T]) ChunkIter() *ChunkIterator[T] {
	return &ChunkIterator[T]{
		pos:  -1,
		pipe: cl,
	}
}

// ValueIterator 的方法
func (it *ValueIterator[T]) Next() bool {
	// 先增加位置
	it.pos++
	return it.pos < it.pipe.size()
}

func (it *ValueIterator[T]) V() T {
	ret, _ := it.pipe.Get(it.pos)
	return ret
}

// ChunkIterator 的方法
func (it *ChunkIterator[T]) Next() bool {
	it.pos++
	return it.pos < len(it.pipe.list)
}

func (it *ChunkIterator[T]) V() []T {
	if it.pos < len(it.pipe.list) && it.pos >= 0 {
		return it.pipe.list[it.pos].val
	}
	var zero []T
	return zero
}

// 當使用完切片後，應該調用這些方法將切片放回 pool
func (cl *ChunkPipe[T]) PutValueSlice(slice []T) {
	if cap(slice) > 0 {
		slice = slice[:0]
		cl.valuePool.Put(&slice)
	}
}

func (cl *ChunkPipe[T]) PutChunkSlice(slice [][]T) {
	if cap(slice) > 0 {
		slice = slice[:0]
		cl.chunkPool.Put(&slice)
	}
}
