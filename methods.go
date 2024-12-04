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

	dataLen := len(data)
	newData := make([]T, dataLen)

	// 使用更快的記憶體複製
	var zero T
	elemSize := unsafe.Sizeof(zero)
	srcPtr := unsafe.Pointer(&data[0])
	dstPtr := unsafe.Pointer(&newData[0])
	totalSize := uintptr(dataLen) * elemSize

	if dataLen <= 32 {
		// 小數據快速路徑
		// 直接使用 uint64 複製
		for i := uintptr(0); i < totalSize; i += 8 {
			*(*uint64)(unsafe.Add(dstPtr, i)) = *(*uint64)(unsafe.Add(srcPtr, i))
		}
	} else {
		// 128字節對齊複製
		aligned128 := totalSize &^ 127
		for i := uintptr(0); i < aligned128; i += 128 {
			// 展開循環，一次複製 128 字節
			*(*[16]uint64)(unsafe.Add(dstPtr, i)) = *(*[16]uint64)(unsafe.Add(srcPtr, i))
		}

		// 64字節對齊複製
		aligned64 := totalSize &^ 63
		for i := aligned128; i < aligned64; i += 64 {
			*(*[8]uint64)(unsafe.Add(dstPtr, i)) = *(*[8]uint64)(unsafe.Add(srcPtr, i))
		}

		// 8字節對齊複製
		aligned8 := totalSize &^ 7
		for i := aligned64; i < aligned8; i += 8 {
			*(*uint64)(unsafe.Add(dstPtr, i)) = *(*uint64)(unsafe.Add(srcPtr, i))
		}

		// 複製剩餘字節
		for i := aligned8; i < totalSize; i++ {
			*(*uint8)(unsafe.Add(dstPtr, i)) = *(*uint8)(unsafe.Add(srcPtr, i))
		}
	}

	block := &Chunk[T]{
		data:   dstPtr,
		size:   dataLen,
		offset: 0,
	}

	// 快速路徑
	tail := cl.tail
	if tail == nil {
		cl.head = block
		cl.tail = block
		cl.totalSize = dataLen
		cl.validSize = dataLen
		return cl
	}

	block.prev = tail
	tail.next = block
	cl.tail = block
	cl.totalSize += dataLen
	cl.validSize += dataLen
	return cl
}

func (cl *ChunkPipe[T]) insertBlockToTree(block *Chunk[T]) {
	if block == nil {
		return
	}

	// 快取常用計算結果
	blockSize := block.size
	validSize := blockSize - block.offset

	newNode := &TreeNode[T]{
		sum:       blockSize,
		validSize: validSize,
		blockAddr: unsafe.Pointer(block),
	}

	// 快取根節點
	root := cl.root
	if root == nil {
		cl.root = newNode
		return
	}

	// 使用局部變數追蹤路徑
	current := root
	for {
		// 更新節點統計
		current.sum += blockSize
		current.validSize += validSize

		// 選擇插入路徑
		if current.left == nil {
			current.left = newNode
			return
		}

		if current.right == nil {
			current.right = newNode
			return
		}

		// 平衡樹
		if current.left.sum <= current.right.sum {
			current = current.left
		} else {
			current = current.right
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

	// 快取常用變數
	head := cl.head
	validSize := cl.validSize
	if head == nil || validSize == 0 {
		return nil, false
	}

	// 快取 size 和 offset
	size := head.size
	offset := head.offset
	validCount := size - offset

	if validCount <= 0 {
		// 移除空塊
		next := head.next
		cl.head = next
		if next != nil {
			next.prev = nil
		} else {
			cl.tail = nil
		}
		return nil, false
	}

	// 創建新的切片並安全複製數據
	newData := make([]T, validCount)
	if head.data != nil {
		src := unsafe.Slice((*T)(head.data), size)
		copy(newData, src[offset:])
	}

	// 更新鏈表
	next := head.next
	cl.head = next
	if next != nil {
		next.prev = nil
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

	// 快取常用變數
	tail := cl.tail
	validSize := cl.validSize
	if tail == nil || validSize == 0 {
		return nil, false
	}

	// 快取 size 和 offset
	size := tail.size
	offset := tail.offset
	validCount := size - offset

	if validCount <= 0 {
		// 移除塊
		prev := tail.prev
		cl.tail = prev
		if prev != nil {
			prev.next = nil
		} else {
			cl.head = nil
		}
		return nil, false
	}

	// 創建新的切片並安全複製數據
	newData := make([]T, validCount)
	if tail.data != nil {
		src := unsafe.Slice((*T)(tail.data), size)
		copy(newData, src[offset:])
	}

	// 更新鏈表
	prev := tail.prev
	cl.tail = prev
	if prev != nil {
		prev.next = nil
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

	head := cl.head
	if head == nil || cl.validSize == 0 {
		return zero, false
	}

	// 直接使用指針操作
	elemSize := unsafe.Sizeof(zero)
	ptr := unsafe.Add(head.data, uintptr(head.offset)*elemSize)
	value := *(*T)(ptr)

	// 使用位運算更新
	head.offset++
	cl.validSize--
	cl.totalSize--

	// 使用更多的位運算和預取
	if i := uintptr(head.offset); (i | uintptr(head.size)) >= uintptr(head.size) {
		// 預取下一個塊
		if next := head.next; next != nil {
			_ = next.data
		}
		// 快速路徑
		next := head.next
		if next != nil {
			next.prev = nil
			cl.head = next
		} else {
			cl.head = nil
			cl.tail = nil
		}
		// 清理指針
		head.next = nil
		head.prev = nil
	}

	return value, true
}

// 抽取共用邏輯
func (cl *ChunkPipe[T]) removeHead() {
	// 快取常用變數
	head := cl.head
	if head == nil {
		return
	}

	// 快取 next 指針
	next := head.next
	cl.head = next

	// 更新指針關係
	if next != nil {
		next.prev = nil
	} else {
		cl.tail = nil
	}

	// 清理原節點的指針
	head.next = nil
	head.prev = nil
}

// 新增尾部移除方法
func (cl *ChunkPipe[T]) removeTail() {
	// 快取常用變數
	tail := cl.tail
	if tail == nil {
		return
	}

	// 快取 prev 指針
	prev := tail.prev
	cl.tail = prev

	// 更新指針關係
	if prev != nil {
		prev.next = nil
	} else {
		cl.head = nil
	}

	// 清理原節點的指針
	tail.next = nil
	tail.prev = nil
}

func (cl *ChunkPipe[T]) PopEnd() (T, bool) {
	var zero T
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// 快取常用變數
	tail := cl.tail
	validSize := cl.validSize
	if tail == nil || validSize == 0 {
		return zero, false
	}

	// 快取 size 和 offset
	size := tail.size
	offset := tail.offset

	// 使用指針計算
	ptr := unsafe.Add(tail.data,
		uintptr(size-1)*unsafe.Sizeof(zero))
	value := *(*T)(ptr)

	size--
	tail.size = size
	cl.validSize--
	cl.totalSize--

	if size <= offset {
		prev := tail.prev
		cl.tail = prev
		if prev != nil {
			prev.next = nil
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

		// 快取常用變數
		head := cl.head
		if head == nil {
			return
		}

		// 優化批次大小
		const (
			minBatchSize = 1024
			maxBatchSize = 8192
		)

		// 使用局部變數追蹤
		for current := head; current != nil; current = current.next {
			// 快取塊屬性
			size := current.size
			offset := current.offset
			if size <= offset {
				continue
			}

			// 計算有效數據小
			validCount := size - offset
			batchSize := validCount
			if batchSize > maxBatchSize {
				batchSize = maxBatchSize
			} else if batchSize < minBatchSize && current.next != nil {
				continue // 小塊等待合併
			}

			// 創建結果切片
			result := make([]T, batchSize)
			src := unsafe.Add(current.data,
				uintptr(offset)*unsafe.Sizeof(result[0]))

			// 使用 copy 而不是 typedmemmove
			srcSlice := unsafe.Slice((*T)(src), batchSize)
			copy(result, srcSlice)

			ch <- result
		}
	}()
	return ch
}

// Range 返回一個支持 for range 的迭代器
func (cl *ChunkPipe[T]) Range() []T {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	validSize := cl.validSize
	if validSize == 0 {
		return nil
	}

	result := make([]T, validSize)
	var zero T
	elemSize := unsafe.Sizeof(zero)
	dstPtr := unsafe.Pointer(&result[0])
	pos := uintptr(0)

	// 使用更大的批次大小
	for current := cl.head; current != nil; current = current.next {
		if n := current.size - current.offset; n > 0 {
			srcPtr := unsafe.Add(current.data, uintptr(current.offset)*elemSize)
			copySize := uintptr(n) * elemSize

			// 256字節對齊複製
			aligned256 := copySize &^ 255
			for i := uintptr(0); i < aligned256; i += 256 {
				*(*[32]uint64)(unsafe.Add(dstPtr, pos+i)) = *(*[32]uint64)(unsafe.Add(srcPtr, i))
			}

			// 128字節對齊複製
			aligned128 := copySize &^ 127
			for i := aligned256; i < aligned128; i += 128 {
				*(*[16]uint64)(unsafe.Add(dstPtr, pos+i)) = *(*[16]uint64)(unsafe.Add(srcPtr, i))
			}

			// 64字節對齊複製
			aligned64 := copySize &^ 63
			for i := aligned128; i < aligned64; i += 64 {
				*(*[8]uint64)(unsafe.Add(dstPtr, pos+i)) = *(*[8]uint64)(unsafe.Add(srcPtr, i))
			}

			// 8字節對齊複製
			aligned8 := copySize &^ 7
			for i := aligned64; i < aligned8; i += 8 {
				*(*uint64)(unsafe.Add(dstPtr, pos+i)) = *(*uint64)(unsafe.Add(srcPtr, i))
			}

			// 複製剩餘字節
			for i := aligned8; i < copySize; i++ {
				*(*uint8)(unsafe.Add(dstPtr, pos+i)) = *(*uint8)(unsafe.Add(srcPtr, i))
			}

			pos += copySize
		}
	}

	return result
}

// RangeValues 提一個優化的類型安全遍歷接口
func (cl *ChunkPipe[T]) RangeValues(fn func(T) bool) {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	head := cl.head
	if head == nil {
		return
	}

	var zero T
	elemSize := unsafe.Sizeof(zero)

	// 使用 16 字節批次處理
	batchSize := uintptr(128) / unsafe.Sizeof(zero)

	for current := head; current != nil; current = current.next {
		if current.size <= current.offset {
			continue
		}

		base := current.data
		offset := uintptr(current.offset)
		size := uintptr(current.size)

		// 批量處理
		for i := offset; i+batchSize <= size; i += batchSize {
			ptr := unsafe.Add(base, i*elemSize)
			// 展開循環
			for j := uintptr(0); j < batchSize; j++ {
				if !fn(*(*T)(unsafe.Add(ptr, j*elemSize))) {
					return
				}
			}
		}

		// 處理剩餘元素
		for i := size - (size-offset)%batchSize; i < size; i++ {
			if !fn(*(*T)(unsafe.Add(base, i*elemSize))) {
				return
			}
		}
	}
}
