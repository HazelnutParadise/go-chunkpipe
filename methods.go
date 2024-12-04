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

	// 直接分配並手動對齊複製
	dataLen := len(data)
	newData := make([]T, dataLen)

	// 手動對齊複製
	var zero T
	elemSize := unsafe.Sizeof(zero)
	srcPtr := unsafe.Pointer(&data[0])
	dstPtr := unsafe.Pointer(&newData[0])

	// 確保 8 字節對齊
	alignedSize := (uintptr(dataLen) * elemSize) &^ 7
	for i := uintptr(0); i < alignedSize; i += 8 {
		*(*uint64)(unsafe.Add(dstPtr, i)) = *(*uint64)(unsafe.Add(srcPtr, i))
	}

	// 複製剩餘字節
	for i := alignedSize; i < uintptr(dataLen)*elemSize; i++ {
		*(*uint8)(unsafe.Add(dstPtr, i)) = *(*uint8)(unsafe.Add(srcPtr, i))
	}

	// 使用局部變數減少結構體訪問
	tail := cl.tail
	block := &Chunk[T]{
		data:   dstPtr,
		size:   dataLen,
		offset: 0,
	}

	// 快速路徑：空鏈表
	if tail == nil {
		cl.head = block
		cl.tail = block

		cl.totalSize = dataLen
		cl.validSize = dataLen
		return cl
	}

	// 快速路徑：追加到尾部
	block.prev = tail
	tail.next = block
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

	// 直接讀取
	elemSize := unsafe.Sizeof(zero)
	ptr := unsafe.Add(head.data, uintptr(head.offset)*elemSize)
	value := *(*T)(ptr)

	// 快速更新
	head.offset++
	cl.validSize--
	cl.totalSize--

	// 快速檢查塊狀態
	if head.offset >= head.size {
		next := head.next
		if next != nil {
			next.prev = nil
			cl.head = next
		} else {
			cl.head = nil
			cl.tail = nil
		}
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

	if cl.validSize == 0 {
		return nil
	}

	// 一次性分配
	result := make([]T, cl.validSize)
	pos := 0

	// 快速複製
	for current := cl.head; current != nil; current = current.next {
		if n := current.size - current.offset; n > 0 {
			copy(result[pos:pos+n],
				unsafe.Slice((*T)(current.data), current.size)[current.offset:])
			pos += n
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

	const batchSize = 32
	var zero T
	elemSize := unsafe.Sizeof(zero)

	for current := head; current != nil; current = current.next {
		if current.size <= current.offset {
			continue
		}

		// 使用指針運算替代切片
		base := current.data
		end := unsafe.Add(base, uintptr(current.size)*elemSize)
		ptr := unsafe.Add(base, uintptr(current.offset)*elemSize)

		for ; uintptr(ptr) < uintptr(end); ptr = unsafe.Add(ptr, elemSize) {
			if !fn(*(*T)(ptr)) {
				return
			}
		}
	}
}
