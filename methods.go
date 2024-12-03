package chunkpipe

import "unsafe"

// 插入數據到 ChunkPipe，支援泛型和鏈式呼叫
func (cl *ChunkPipe[T]) Push(data []T) *ChunkPipe[T] {
	// 小數據快速路徑
	if len(data) <= 64 {
		cl.mu.Lock()
		if cl.tail != nil && cl.tail.size-cl.tail.offset < 64 {
			// 直接追加到現有塊
			ptr := unsafe.Add(cl.tail.data, uintptr(cl.tail.size)*unsafe.Sizeof(data[0]))
			copy(unsafe.Slice((*T)(ptr), len(data)), data)
			cl.tail.size += len(data)
			cl.totalSize += len(data)
			cl.validSize += len(data)
			cl.mu.Unlock()
			return cl
		}
		cl.mu.Unlock()
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	block := &Chunk[T]{
		data:   unsafe.Pointer(&data[0]),
		size:   len(data),
		offset: 0,
	}

	if cl.tail != nil {
		cl.tail.next = block
		block.prev = cl.tail
	} else {
		cl.head = block
	}
	cl.tail = block

	cl.insertBlockToTree(block)
	cl.totalSize += len(data)
	cl.validSize += len(data)
	return cl
}

func (cl *ChunkPipe[T]) insertBlockToTree(block *Chunk[T]) {
	newNode := &TreeNode[T]{
		sum:       block.size,
		validSize: block.size - block.offset,
		blockAddr: unsafe.Pointer(block),
	}

	if cl.root == nil {
		cl.root = newNode
		return
	}

	current := cl.root
	for {
		current.sum += block.size
		current.validSize += (block.size - block.offset)
		if current.left == nil {
			current.left = newNode
			return
		} else if current.right == nil {
			current.right = newNode
			return
		} else {
			if current.left.sum <= current.right.sum {
				current = current.left
			} else {
				current = current.right
			}
		}
	}
}

func (cl *ChunkPipe[T]) Get(index int) (T, bool) {
	var zero T

	cl.mu.RLock()
	defer cl.mu.RUnlock()

	if index < 0 || index >= cl.validSize {
		return zero, false
	}

	current := cl.head
	remainingIndex := index

	for current != nil {
		validCount := current.size - current.offset
		if remainingIndex < validCount {
			ptr := unsafe.Add(current.data, uintptr(current.offset+remainingIndex)*unsafe.Sizeof(*(*T)(current.data)))
			return *(*T)(ptr), true
		}
		remainingIndex -= validCount
		current = current.next
	}

	return zero, false
}

// 從頭部彈出數據
func (cl *ChunkPipe[T]) PopChunkFront() ([]T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.head == nil {
		return nil, false
	}

	block := cl.head
	cl.head = block.next
	if cl.head != nil {
		cl.head.prev = nil
	} else {
		cl.tail = nil
	}

	validCount := block.size - block.offset
	cl.totalSize -= validCount
	cl.validSize -= validCount

	// 使用指針計算創建新的切片
	newData := make([]T, validCount)
	for i := 0; i < validCount; i++ {
		ptr := unsafe.Add(block.data, uintptr(block.offset+i)*unsafe.Sizeof(*(*T)(block.data)))
		newData[i] = *(*T)(ptr)
	}
	return newData, true
}

// 從尾部彈出數據
func (cl *ChunkPipe[T]) PopChunkEnd() ([]T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.tail == nil {
		return nil, false
	}

	block := cl.tail
	cl.tail = block.prev
	if cl.tail != nil {
		cl.tail.next = nil
	} else {
		cl.head = nil
	}

	validCount := block.size - block.offset
	cl.totalSize -= validCount
	cl.validSize -= validCount

	data := unsafe.Slice((*T)(block.data), block.size)
	return data[block.offset:block.size], true
}

func (cl *ChunkPipe[T]) PopFront() (T, bool) {
	var zero T
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.head == nil || cl.validSize == 0 {
		return zero, false
	}

	block := cl.head
	if block.offset >= block.size {
		cl.head = block.next
		if cl.head != nil {
			cl.head.prev = nil
		} else {
			cl.tail = nil
		}
		return zero, false
	}

	ptr := unsafe.Add(block.data, uintptr(block.offset)*unsafe.Sizeof(*(*T)(block.data)))
	value := *(*T)(ptr)

	block.offset++
	cl.validSize--
	cl.totalSize--

	if block.offset >= block.size {
		cl.head = block.next
		if cl.head != nil {
			cl.head.prev = nil
		} else {
			cl.tail = nil
		}
	}

	return value, true
}

func (cl *ChunkPipe[T]) PopEnd() (T, bool) {
	var zero T
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.tail == nil || cl.validSize == 0 {
		return zero, false
	}

	block := cl.tail
	// 使用指針計算
	ptr := unsafe.Add(block.data, uintptr(block.size-1)*unsafe.Sizeof(*(*T)(block.data)))
	value := *(*T)(ptr)

	block.size--
	cl.validSize--
	cl.totalSize--

	if block.size <= block.offset {
		cl.tail = block.prev
		if cl.tail != nil {
			cl.tail.next = nil
		} else {
			cl.head = nil
		}
	}

	return value, true
}

// 重命名原來的 Range 為 RangeChunk
func (cl *ChunkPipe[T]) RangeChunk() <-chan []T {
	ch := make(chan []T)
	go func() {
		cl.mu.RLock()
		defer cl.mu.RUnlock()

		current := cl.head
		for current != nil {
			if current.offset < current.size {
				validCount := current.size - current.offset
				newData := make([]T, validCount)
				// 使用指針直接複製數據
				for i := 0; i < validCount; i++ {
					ptr := unsafe.Add(current.data, uintptr(current.offset+i)*unsafe.Sizeof(*(*T)(current.data)))
					newData[i] = *(*T)(ptr)
				}
				ch <- newData
			}
			current = current.next
		}
		close(ch)
	}()
	return ch
}

// 新增高性能的單元素 Range 方法
func (cl *ChunkPipe[T]) Range() <-chan T {
	ch := make(chan T, 8192)
	go func() {
		cl.mu.RLock()
		defer cl.mu.RUnlock()

		current := cl.head
		for current != nil {
			if current.size > current.offset {
				basePtr := current.data
				for i := current.offset; i < current.size; i++ {
					ptr := unsafe.Add(basePtr, uintptr(i)*unsafe.Sizeof(*(*T)(basePtr)))
					ch <- *(*T)(ptr)
				}
			}
			current = current.next
		}
		close(ch)
	}()
	return ch
}
