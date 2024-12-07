package chunkpipe

// 插入數據到 ChunkPipe，支援泛型和鏈式呼叫
func (cl *ChunkPipe[T]) Push(data []T) *ChunkPipe[T] {
	dataLen := len(data)

	if dataLen == 0 {
		cl.mu.RLock()
		defer cl.mu.RUnlock()
		return cl
	}

	cl.mu.Lock()
	off := cl.offset
	listPtr := cl.listPool.Get().(*[]offset[T])
	defer func() {
		cl.listPool.Put(listPtr)
		cl.mu.Unlock()
	}()
	list := *listPtr
	listLen := len(list)

	if listLen != 0 {
		off = list[listLen-1].off
	}

	*listPtr = append(list, offset[T]{
		val: data,
		off: off + dataLen,
	})

	return cl
}

func (cl *ChunkPipe[T]) Get(index int) (T, bool) {
	var zero T
	cl.mu.Lock()
	listPtr := cl.listPool.Get().(*[]offset[T])
	defer func() {
		cl.listPool.Put(listPtr)
		cl.mu.Unlock()
	}()
	list := *listPtr
	listLen := len(list)

	if listLen == 0 || index < 0 {
		return zero, false
	}

	target := index + cl.offset
	l := 0
	r := listLen - 1

	if target >= list[r].off {
		return zero, false
	}

	if list[l].off > target {
		off := list[0]
		val := off.val
		target = len(val) - (off.off - target)
		return val[target], true
	}

	for r-l > 1 {
		m := (r + l) >> 1
		if list[m].off > target {
			r = m
		} else {
			l = m
		}
	}

	off := list[r]
	val := off.val
	target = len(val) - (off.off - target)
	return val[target], true
}

// 從頭部彈出數據
func (cl *ChunkPipe[T]) PopChunkFront() ([]T, bool) {
	cl.mu.Lock()

	listPtr := cl.listPool.Get().(*[]offset[T])
	defer func() {
		cl.listPool.Put(listPtr)
		cl.mu.Unlock()
	}()
	list := *listPtr
	listLen := len(list)
	if listLen > 0 {
		cl.offset = list[0].off
		ret := list[0].val
		*listPtr = list[1:]
		return ret, true
	}
	return nil, false
}

// 從尾部彈出數據
func (cl *ChunkPipe[T]) PopChunkEnd() ([]T, bool) {
	cl.mu.Lock()

	listPtr := cl.listPool.Get().(*[]offset[T])
	defer func() {
		cl.listPool.Put(listPtr)
		cl.mu.Unlock()
	}()
	list := *listPtr
	listLen := len(list)
	listLenMinusOne := listLen - 1

	if listLen > 0 {
		ret := list[listLenMinusOne].val
		*listPtr = list[:listLenMinusOne]
		return ret, true
	}
	return nil, false
}

func (cl *ChunkPipe[T]) PopFront() (T, bool) {
	cl.mu.Lock()

	listPtr := cl.listPool.Get().(*[]offset[T])
	defer func() {
		cl.listPool.Put(listPtr)
		cl.mu.Unlock()
	}()
	list := *listPtr
	listLen := len(list)

	if listLen > 0 {
		val := list[0].val
		ret := val[0]
		val = val[1:]
		list[0].val = val
		cl.offset++
		if len(val) == 0 {
			*listPtr = list[1:]
		}
		return ret, true
	}
	var ret T
	return ret, false
}

// 從尾部彈出數據
func (cl *ChunkPipe[T]) PopEnd() (T, bool) {
	cl.mu.Lock()

	listPtr := cl.listPool.Get().(*[]offset[T])
	defer func() {
		cl.listPool.Put(listPtr)
		cl.mu.Unlock()
	}()
	list := *listPtr
	listLen := len(list)

	var ret T
	if listLen == 0 {
		return ret, false
	}

	listLenMinusOne := listLen - 1
	val := list[listLenMinusOne].val
	valLen := len(val)

	if valLen == 0 {
		// 如果當前塊為空，移除整�塊
		*listPtr = list[:listLenMinusOne]
		return ret, false
	}

	valLenMinusOne := valLen - 1
	ret = val[valLenMinusOne]
	val = val[:valLenMinusOne]
	list[listLenMinusOne].val = val
	list[listLenMinusOne].off--

	if valLen == 1 {
		// 如果這是塊中的最後一個元素，移除整個塊
		*listPtr = list[:listLenMinusOne]
	}

	return ret, true
}

// ValueSlice 返回所有值的切片
func (cl *ChunkPipe[T]) ValueSlice() []T {
	cl.mu.Lock()

	listPtr := cl.listPool.Get().(*[]offset[T])
	defer func() {
		cl.listPool.Put(listPtr)
		cl.mu.Unlock()
	}()
	list := *listPtr
	if len(list) == 0 {
		return []T{}
	}

	size := list[len(list)-1].off - cl.offset
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
	for i := range list {
		for _, v := range list[i].val {
			ret[k] = v
			k++
		}
	}
	return ret
}

// ChunkSlice 返回所有數據塊的切片
func (cl *ChunkPipe[T]) ChunkSlice() [][]T {
	cl.mu.Lock()

	listPtr := cl.listPool.Get().(*[]offset[T])
	defer func() {
		cl.listPool.Put(listPtr)
		cl.mu.Unlock()
	}()
	list := *listPtr
	listLen := len(list)

	if listLen == 0 {
		return [][]T{}
	}

	// 從 cl.chunkSlicePool 中獲取切片
	slicePtr := cl.chunkSlicePool.Get().(*[][]T)
	ret := *slicePtr
	// 確保切片容量足夠
	if cap(ret) < listLen {
		go func() { cl.chunkSlicePool.Put(slicePtr) }()
		ret = make([][]T, listLen)
	} else {
		ret = ret[:listLen]
	}

	for i := range ret {
		ret[i] = list[i].val
	}
	return ret
}

func (cl *ChunkPipe[T]) size() int {
	cl.mu.Lock()

	listPtr := cl.listPool.Get().(*[]offset[T])
	defer func() {
		cl.listPool.Put(listPtr)
		cl.mu.Unlock()
	}()
	list := *listPtr
	listLen := len(list)

	if listLen == 0 {
		return 0
	}
	return list[listLen-1].off - cl.offset
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
	listPtr := it.pipe.listPool.Get().(*[]offset[T])
	defer it.pipe.listPool.Put(listPtr)
	list := *listPtr

	return it.pos < len(list)
}

func (it *ChunkIterator[T]) V() []T {
	listPtr := it.pipe.listPool.Get().(*[]offset[T])
	defer it.pipe.listPool.Put(listPtr)
	list := *listPtr

	if it.pos < len(list) && it.pos >= 0 {
		return list[it.pos].val
	}
	var zero []T
	return zero
}
