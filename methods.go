package chunkpipe

import (
	"unsafe"
	_ "unsafe"
)

//go:linkname typedmemmove runtime.typedmemmove
func typedmemmove(typ unsafe.Pointer, dst, src unsafe.Pointer, len int)

// 插入數據到 ChunkPipe，支援泛型和鏈式呼叫
func (cl *ChunkPipe[T]) Push(data []T) *ChunkPipe[T] {
	if len(data) == 0 {
		return cl
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	// 小數據優化
	if len(data) <= 16 {
		if cl.tail != nil && cl.tail.size-cl.tail.offset < 32 {
			// 創建新的內存空間
			newSize := cl.tail.size + len(data)
			newData := make([]T, newSize)

			// 複製現有數據
			oldSlice := unsafe.Slice((*T)(cl.tail.data), cl.tail.size)
			copy(newData, oldSlice)

			// 複製新數據
			copy(newData[cl.tail.size:], data)

			// 更新塊信息
			cl.tail.data = unsafe.Pointer(&newData[0])
			cl.tail.size = newSize
			cl.totalSize += len(data)
			cl.validSize += len(data)
			return cl
		}
	}

	// 大數據處理
	newData := make([]T, len(data))
	copy(newData, data)

	block := &Chunk[T]{
		data:   unsafe.Pointer(&newData[0]),
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

	cl.totalSize += len(data)
	cl.validSize += len(data)
	return cl
}

func (cl *ChunkPipe[T]) insertBlockToTree(block *Chunk[T]) {
	if block == nil {
		return
	}

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

	if cl.head == nil || cl.validSize == 0 {
		return nil, false
	}

	block := cl.head
	validCount := block.size - block.offset
	if validCount <= 0 {
		// 移除空塊
		cl.head = block.next
		if cl.head != nil {
			cl.head.prev = nil
		} else {
			cl.tail = nil
		}
		return nil, false
	}

	// 創建新的切片並安全複製數據
	newData := make([]T, validCount)
	if block.data != nil {
		src := unsafe.Slice((*T)(block.data), block.size)
		copy(newData, src[block.offset:])
	}

	// 更新鏈表
	cl.head = block.next
	if cl.head != nil {
		cl.head.prev = nil
	} else {
		cl.tail = nil
	}

	// 更新計數
	cl.totalSize -= validCount
	cl.validSize -= validCount

	return newData, true
}

// 從尾部彈出數據
func (cl *ChunkPipe[T]) PopChunkEnd() ([]T, bool) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.tail == nil || cl.validSize == 0 {
		return nil, false
	}

	block := cl.tail
	validCount := block.size - block.offset
	if validCount <= 0 {
		// 移除塊
		cl.tail = block.prev
		if cl.tail != nil {
			cl.tail.next = nil
		} else {
			cl.head = nil
		}
		return nil, false
	}

	// 創建新的切片並安全複製數
	newData := make([]T, validCount)
	if block.data != nil {
		src := unsafe.Slice((*T)(block.data), block.size)
		copy(newData, src[block.offset:])
	}

	// 更新鏈表
	cl.tail = block.prev
	if cl.tail != nil {
		cl.tail.next = nil
	} else {
		cl.head = nil
	}

	// 更新計數
	cl.totalSize -= validCount
	cl.validSize -= validCount

	return newData, true
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
		cl.removeHead()
		return zero, false
	}

	// 批量讀取優化
	ptr := unsafe.Add(block.data, uintptr(block.offset)*unsafe.Sizeof(*(*T)(block.data)))
	value := *(*T)(ptr)

	block.offset++
	cl.validSize--
	cl.totalSize--

	// 優化塊移除邏輯
	if block.offset >= block.size-32 { // 提高閾值
		if block.offset >= block.size {
			cl.removeHead()
		}
	}

	return value, true
}

// 抽取共用邏輯
func (cl *ChunkPipe[T]) removeHead() {
	cl.head = cl.head.next
	if cl.head != nil {
		cl.head.prev = nil
	} else {
		cl.tail = nil
	}
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
	ch := make(chan []T, 256) // 更大的緩衝區
	go func() {
		cl.mu.RLock()
		defer cl.mu.RUnlock()

		// 預分配一個大的切片
		buffer := make([]T, 0, 1024)

		current := cl.head
		for current != nil {
			if current.offset < current.size {
				// 直接使用原始數據
				data := unsafe.Slice((*T)(current.data), current.size)
				validData := data[current.offset:]

				// 如果緩衝區足夠，直接追加
				if len(buffer)+len(validData) <= cap(buffer) {
					buffer = append(buffer, validData...)
				} else {
					// 發送當前緩衝區
					if len(buffer) > 0 {
						ch <- buffer
						buffer = make([]T, 0, 1024)
					}
					// 直接發送大塊數據
					ch <- validData
				}
			}
			current = current.next
		}

		// 發送剩餘的數據
		if len(buffer) > 0 {
			ch <- buffer
		}

		close(ch)
	}()
	return ch
}

// Range 返回一個支持 for range 的迭代器
func (cl *ChunkPipe[T]) Range() []T {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	totalSize := cl.validSize
	if totalSize == 0 {
		return nil
	}

	// 預分配精確大小的切片
	result := make([]T, 0, totalSize)

	// 使用更大的批次大小
	const batchSize = 64
	buffer := make([]T, batchSize)

	current := cl.head
	for current != nil {
		if current.size > current.offset {
			basePtr := unsafe.Add(current.data,
				uintptr(current.offset)*unsafe.Sizeof(*(*T)(current.data)))
			remaining := current.size - current.offset

			// 批量複製
			for i := 0; i < remaining; i += batchSize {
				n := batchSize
				if i+n > remaining {
					n = remaining - i
				}
				src := unsafe.Slice((*T)(unsafe.Add(basePtr, uintptr(i)*unsafe.Sizeof(*(*T)(basePtr)))), n)
				copy(buffer[:n], src)
				result = append(result, buffer[:n]...)
			}
		}
		current = current.next
	}

	return result
}

// RangeValues 提一個優化的類型安全遍歷接口
func (cl *ChunkPipe[T]) RangeValues(fn func(T) bool) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	current := cl.head
	for current != nil {
		if current.size > current.offset {
			// 創建一次性切片視圖
			slice := unsafe.Slice((*T)(current.data), current.size)

			// 使用 CPU 友好的步長
			const batchSize = 16

			// 主循環：批量處理
			i := current.offset
			for ; i+batchSize <= current.size; i += batchSize {
				// 預取下一批數據
				if i+batchSize*2 <= current.size {
					_ = slice[i+batchSize]
				}

				// 展開循環以提高指令級並行性
				if !fn(slice[i]) || !fn(slice[i+1]) ||
					!fn(slice[i+2]) || !fn(slice[i+3]) ||
					!fn(slice[i+4]) || !fn(slice[i+5]) ||
					!fn(slice[i+6]) || !fn(slice[i+7]) ||
					!fn(slice[i+8]) || !fn(slice[i+9]) ||
					!fn(slice[i+10]) || !fn(slice[i+11]) ||
					!fn(slice[i+12]) || !fn(slice[i+13]) ||
					!fn(slice[i+14]) || !fn(slice[i+15]) {
					return
				}
			}

			// 處理剩餘元素
			for ; i < current.size; i++ {
				if !fn(slice[i]) {
					return
				}
			}
		}
		current = current.next
	}
}
