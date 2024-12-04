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

	// 快取長度避免重複計算
	dataLen := len(data)

	// 直接創建新塊並複製數據
	newData := make([]T, dataLen)
	copy(newData, data)

	block := &Chunk[T]{
		data:   unsafe.Pointer(&newData[0]),
		size:   dataLen,
		offset: 0,
	}

	// 快取 tail 避免多次訪問
	tail := cl.tail
	if tail != nil {
		tail.next = block
		block.prev = tail
	} else {
		cl.head = block
	}
	cl.tail = block

	// 一次性更新計數
	cl.totalSize += dataLen
	cl.validSize += dataLen
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

	// 使用快速路徑
	if cl.head != nil && index < cl.head.size-cl.head.offset {
		ptr := unsafe.Add(cl.head.data,
			uintptr(cl.head.offset+index)*unsafe.Sizeof(zero))
		return *(*T)(ptr), true
	}

	// 慢路徑
	current := cl.head
	remainingIndex := index
	for current != nil {
		validCount := current.size - current.offset
		if remainingIndex < validCount {
			ptr := unsafe.Add(current.data,
				uintptr(current.offset+remainingIndex)*unsafe.Sizeof(zero))
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

	// 新計數
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

	// 快取所有常用變數
	head := cl.head
	validSize := cl.validSize
	if head == nil || validSize == 0 {
		return zero, false
	}

	// 快取 offset 和 size
	offset := head.offset
	size := head.size

	// 快速路徑
	value := *(*T)(unsafe.Add(head.data,
		uintptr(offset)*unsafe.Sizeof(zero)))

	offset++
	head.offset = offset
	cl.validSize--
	cl.totalSize--

	// 如果當前塊已空，移除它
	if offset >= size {
		next := head.next
		cl.head = next
		if next != nil {
			next.prev = nil
		} else {
			cl.tail = nil
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
	ch := make(chan []T, 256)
	go func() {
		cl.mu.RLock()
		defer cl.mu.RUnlock()
		defer close(ch)

		const maxBatchSize = 4096
		current := cl.head
		for current != nil {
			if current.size > current.offset {
				validCount := current.size - current.offset
				if validCount > maxBatchSize {
					validCount = maxBatchSize
				}

				result := make([]T, validCount)
				typedmemmove(unsafe.Pointer(&result[0]),
					unsafe.Pointer(&result[0]),
					unsafe.Add(current.data,
						uintptr(current.offset)*unsafe.Sizeof(result[0])),
					validCount)

				ch <- result
			}
			current = current.next
		}
	}()
	return ch
}

// Range 返回一個支持 for range 的迭代器
func (cl *ChunkPipe[T]) Range() []T {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	// 快取所有常用變數
	validSize := cl.validSize
	head := cl.head
	if validSize == 0 || head == nil {
		return nil
	}

	// 預分配結果切片
	result := make([]T, validSize)
	pos := 0

	// 使用批量複製
	for current := head; current != nil; current = current.next {
		if current.size > current.offset {
			validCount := current.size - current.offset
			src := unsafe.Slice((*T)(current.data), current.size)
			dst := result[pos : pos+validCount]
			copy(dst, src[current.offset:])
			pos += validCount
		}
	}

	return result
}

// RangeValues 提一個優化的類型安全遍歷接口
func (cl *ChunkPipe[T]) RangeValues(fn func(T) bool) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	// 快取常用變數
	head := cl.head
	if head == nil {
		return
	}

	const batchSize = 32 // 增加批次大小

	for current := head; current != nil; current = current.next {
		if current.size <= current.offset {
			continue
		}

		// 創建一次性切片視圖
		slice := unsafe.Slice((*T)(current.data), current.size)
		offset := current.offset
		size := current.size

		// 主循環：批量處理
		for i := offset; i+batchSize <= size; i += batchSize {
			// 預取下一批數據
			if i+batchSize*2 <= size {
				_ = slice[i+batchSize]
			}

			// 展開循環
			for j := 0; j < batchSize; j++ {
				if !fn(slice[i+j]) {
					return
				}
			}
		}

		// 處理剩餘元素
		for i := size - (size-offset)%batchSize; i < size; i++ {
			if !fn(slice[i]) {
				return
			}
		}
	}
}
