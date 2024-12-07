package chunkpipe

// 插入數據到 ChunkPipe，支援泛型和鏈式呼叫
func (cl *ChunkPipe[T]) Push(data []T) *ChunkPipe[T] {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if len(data) == 0 {
		return cl
	}

	off := cl.offset
	list := cl.listPool.Get().(*[]offset[T])
	defer cl.listPool.Put(list)

	if len(*list) != 0 {
		off = (*list)[len(*list)-1].off
	}

	*list = append(*list, offset[T]{
		val: data,
		off: off + len(data),
	})

	return cl
}

func (cl *ChunkPipe[T]) Get(index int) (T, bool) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	var zero T
	list := cl.listPool.Get().(*[]offset[T])
	defer cl.listPool.Put(list)

	if len(*list) == 0 || index < 0 {
		return zero, false
	}

	target := index + cl.offset
	l := 0
	r := len(*list) - 1

	if target >= (*list)[r].off {
		return zero, false
	}

	if (*list)[l].off > target {
		off := (*list)[0]
		val := off.val
		target = len(val) - (off.off - target)
		return val[target], true
	}

	for r-l > 1 {
		m := (r + l) >> 1
		if (*list)[m].off > target {
			r = m
		} else {
			l = m
		}
	}

	off := (*list)[r]
	val := off.val
	target = len(val) - (off.off - target)
	return val[target], true
}

// 從頭部彈出數據
func (cl *ChunkPipe[T]) PopChunkFront() ([]T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	list := cl.listPool.Get().(*[]offset[T])
	defer cl.listPool.Put(list)

	if len(*list) > 0 {
		cl.offset = (*list)[0].off
		ret := (*list)[0].val
		*list = (*list)[1:]
		return ret, true
	}
	return nil, false
}

// 從尾部彈出數據
func (cl *ChunkPipe[T]) PopChunkEnd() ([]T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	list := cl.listPool.Get().(*[]offset[T])
	defer cl.listPool.Put(list)

	if len(*list) > 0 {
		ret := (*list)[len(*list)-1].val
		*list = (*list)[:len(*list)-1]
		return ret, true
	}
	return nil, false
}

func (cl *ChunkPipe[T]) PopFront() (T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	list := cl.listPool.Get().(*[]offset[T])
	defer cl.listPool.Put(list)

	if len(*list) > 0 {
		val := (*list)[0].val
		ret := val[0]
		val = val[1:]
		(*list)[0].val = val
		cl.offset++
		if len(val) == 0 {
			*list = (*list)[1:]
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

	list := cl.listPool.Get().(*[]offset[T])
	defer cl.listPool.Put(list)

	if len(*list) > 0 {
		val := (*list)[len(*list)-1].val
		ret := val[len(val)-1]
		val = val[:len(val)-1]
		(*list)[len(*list)-1].val = val
		(*list)[len(*list)-1].off--

		if len(val) == 0 {
			// remove the element
			*list = (*list)[:len(*list)-1]
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

	list := cl.listPool.Get().(*[]offset[T])
	defer cl.listPool.Put(list)

	if len(*list) == 0 {
		return []T{}
	}

	size := (*list)[len(*list)-1].off - cl.offset
	// 從 cl.valueSlicePool 中獲取切片
	slicePtr := cl.valueSlicePool.Get().(*[]T)
	ret := *slicePtr
	// 確保切片容量足夠
	if cap(ret) < size {
		go cl.valueSlicePool.Put(slicePtr)
		ret = make([]T, size)
	} else {
		ret = ret[:size]
	}

	k := 0
	for i := range *list {
		for _, v := range (*list)[i].val {
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

	list := cl.listPool.Get().(*[]offset[T])
	defer cl.listPool.Put(list)

	if len(*list) == 0 {
		return [][]T{}
	}

	// 從 cl.chunkSlicePool 中獲取切片
	slicePtr := cl.chunkSlicePool.Get().(*[][]T)
	ret := *slicePtr
	// 確保切片容量足夠
	if cap(ret) < len(*list) {
		go cl.chunkSlicePool.Put(slicePtr)
		ret = make([][]T, len(*list))
	} else {
		ret = ret[:len(*list)]
	}

	for i := range ret {
		ret[i] = (*list)[i].val
	}
	return ret
}

func (cl *ChunkPipe[T]) size() int {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	list := cl.listPool.Get().(*[]offset[T])
	defer cl.listPool.Put(list)

	if len(*list) == 0 {
		return 0
	}
	return (*list)[len(*list)-1].off - cl.offset
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
	list := it.pipe.listPool.Get().(*[]offset[T])
	defer it.pipe.listPool.Put(list)

	return it.pos < len(*list)
}

func (it *ChunkIterator[T]) V() []T {
	list := it.pipe.listPool.Get().(*[]offset[T])
	defer it.pipe.listPool.Put(list)

	if it.pos < len(*list) && it.pos >= 0 {
		return (*list)[it.pos].val
	}
	var zero []T
	return zero
}
