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

	// 先分配記憶體並異步複製數據
	dataLen := len(data)
	newData := make([]T, dataLen)

	var zero T
	elemSize := unsafe.Sizeof(zero)
	srcPtr := unsafe.Pointer(&data[0])
	dstPtr := unsafe.Pointer(&newData[0])
	totalSize := uintptr(dataLen) * elemSize

	// 使用多個 goroutine 並行複製大塊數據
	if totalSize >= 4096 { // 4KB 以上使用並行
		const chunkSize uintptr = 512 // 每個 goroutine 處理 512 字節
		numGoroutines := (totalSize + chunkSize - 1) / chunkSize
		done := make(chan struct{}, numGoroutines)

		for offset := uintptr(0); offset < totalSize; offset += chunkSize {
			go func(off uintptr) {
				size := chunkSize
				if off+size > totalSize {
					size = totalSize - off
				}

				// 手動記憶體複製使用 64 字節批次
				for i := uintptr(0); i < size; i += 64 {
					if i+64 <= size {
						*(*[8]uint64)(unsafe.Add(dstPtr, off+i)) =
							*(*[8]uint64)(unsafe.Add(srcPtr, off+i))
					} else {
						// 處理剩餘字節
						for j := i; j < size; j += 8 {
							*(*uint64)(unsafe.Add(dstPtr, off+j)) =
								*(*uint64)(unsafe.Add(srcPtr, off+j))
						}
					}
				}
				done <- struct{}{}
			}(offset)
		}

		// 等待所有複製完成
		for i := uintptr(0); i < numGoroutines; i++ {
			<-done
		}
	} else {
		// 小數據使用 64 字節批次複製
		aligned64 := totalSize &^ 63
		for i := uintptr(0); i < aligned64; i += 64 {
			*(*[8]uint64)(unsafe.Add(dstPtr, i)) =
				*(*[8]uint64)(unsafe.Add(srcPtr, i))
		}

		// 處理剩餘字節
		for i := aligned64; i < totalSize; i += 8 {
			*(*uint64)(unsafe.Add(dstPtr, i)) =
				*(*uint64)(unsafe.Add(srcPtr, i))
		}
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	block := &Chunk[T]{
		data:   dstPtr,
		size:   int32(dataLen),
		offset: 0,
	}

	// 更新 B+ 樹索引
	if cl.bptree != nil {
		cl.bptree.Insert(uintptr(cl.validSize), dstPtr)
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
	blockSize := int(block.size)
	validSize := int(block.size - block.offset)

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

	// 使用 B+ 樹快速定位
	if cl.bptree != nil {
		// 使用索引位置作為鍵值
		data, offset := cl.bptree.Find(uintptr(index))
		if data != nil {
			ptr := unsafe.Add(data, offset*unsafe.Sizeof(zero))
			return *(*T)(ptr), true
		}
	}

	// 快速路徑：檢查頭部
	if cl.head != nil && int(cl.head.size-cl.head.offset) > index {
		ptr := unsafe.Add(cl.head.data,
			uintptr(int(cl.head.offset)+index)*unsafe.Sizeof(zero))
		return *(*T)(ptr), true
	}

	// 慢路徑：遍歷鏈表
	current := cl.head
	remainingIndex := index
	for current != nil {
		validCount := int(current.size - current.offset)
		if remainingIndex < validCount {
			ptr := unsafe.Add(current.data,
				uintptr(int(current.offset)+remainingIndex)*unsafe.Sizeof(zero))
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
	validCount := int(size - offset)

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
	validCount := int(size - offset)

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

	// 使用指計算
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

	// 使用 B+ 樹進行順序訪問
	if cl.bptree != nil {
		// TODO: 實現 B+ 樹的順序遍歷
		// 這裡需要實現 B+ 樹的葉子節點鏈表遍歷
	}

	// 使用多個 goroutine 行複製
	const chunkSize uintptr = 4096
	for current := cl.head; current != nil; current = current.next {
		if n := current.size - current.offset; n > 0 {
			copySize := uintptr(n) * elemSize
			if copySize >= chunkSize {
				// 大塊數據使用並行複製
				numGoroutines := (copySize + chunkSize - 1) / chunkSize
				done := make(chan struct{}, numGoroutines)

				for offset := uintptr(0); offset < copySize; offset += chunkSize {
					go func(off uintptr) {
						size := chunkSize
						if off+size > copySize {
							size = copySize - off
						}

						srcPtr := unsafe.Add(current.data,
							uintptr(current.offset)*elemSize+off)

						// 使用 64 字節批次複製
						for i := uintptr(0); i < size; i += 64 {
							if i+64 <= size {
								*(*[8]uint64)(unsafe.Add(dstPtr, pos+off+i)) =
									*(*[8]uint64)(unsafe.Add(srcPtr, i))
							} else {
								for j := i; j < size; j += 8 {
									*(*uint64)(unsafe.Add(dstPtr, pos+off+j)) =
										*(*uint64)(unsafe.Add(srcPtr, j))
								}
							}
						}
						done <- struct{}{}
					}(offset)
				}

				for i := uintptr(0); i < numGoroutines; i++ {
					<-done
				}
			} else {
				// 小塊數據直接複製
				srcPtr := unsafe.Add(current.data,
					uintptr(current.offset)*elemSize)

				// 使用 64 字節批次複製
				aligned64 := copySize &^ 63
				for i := uintptr(0); i < aligned64; i += 64 {
					*(*[8]uint64)(unsafe.Add(dstPtr, pos+i)) =
						*(*[8]uint64)(unsafe.Add(srcPtr, i))
				}

				// 處理剩餘字節
				for i := aligned64; i < copySize; i += 8 {
					*(*uint64)(unsafe.Add(dstPtr, pos+i)) =
						*(*uint64)(unsafe.Add(srcPtr, i))
				}
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

// 優化並行複製的閾值和批次大小
const (
	parallelThreshold = 4096 // 4KB
	copyBatchSize     = 512  // 512B
	maxGoroutines     = 32   // 最大 goroutine 數
)

// 使用工作池優化並行複製
var copyWorkerPool = make(chan struct{}, maxGoroutines)

func init() {
	// 初始化工作池
	for i := 0; i < maxGoroutines; i++ {
		copyWorkerPool <- struct{}{}
	}
}

// ValueSlice 返回所有值的切片
func (cl *ChunkPipe[T]) ValueSlice() []T {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	if cl.validSize == 0 {
		return nil
	}

	result := make([]T, cl.validSize)
	var zero T
	elemSize := unsafe.Sizeof(zero)
	dstPtr := unsafe.Pointer(&result[0])
	pos := uintptr(0)

	for current := cl.head; current != nil; current = current.next {
		if n := current.size - current.offset; n > 0 {
			copySize := uintptr(n) * elemSize
			srcPtr := unsafe.Add(current.data,
				uintptr(current.offset)*elemSize)

			// 使用 64 字節批次複製
			aligned64 := copySize &^ 63
			for i := uintptr(0); i < aligned64; i += 64 {
				*(*[8]uint64)(unsafe.Add(dstPtr, pos+i)) =
					*(*[8]uint64)(unsafe.Add(srcPtr, i))
			}

			// 處理剩餘字節
			for i := aligned64; i < copySize; i += 8 {
				*(*uint64)(unsafe.Add(dstPtr, pos+i)) =
					*(*uint64)(unsafe.Add(srcPtr, i))
			}
			pos += copySize
		}
	}

	return result
}

// ChunkSlice 返回所有數據塊的切片
func (cl *ChunkPipe[T]) ChunkSlice() [][]T {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	if cl.validSize == 0 {
		return nil
	}

	chunks := make([][]T, 0)
	for current := cl.head; current != nil; current = current.next {
		if n := current.size - current.offset; n > 0 {
			chunk := make([]T, n)
			src := unsafe.Add(current.data,
				uintptr(current.offset)*unsafe.Sizeof(chunk[0]))
			srcSlice := unsafe.Slice((*T)(src), n)
			copy(chunk, srcSlice)
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// ValueIter 返回值迭代器
func (cl *ChunkPipe[T]) ValueIter() *ValueIterator[T] {
	// 原來的 Values 方法實現
	return &ValueIterator[T]{
		current: cl.head,
		pos:     0,
		pipe:    cl,
	}
}

// ChunkIter 返回塊迭代器
func (cl *ChunkPipe[T]) ChunkIter() *ChunkIterator[T] {
	return &ChunkIterator[T]{
		current: cl.head,
		pipe:    cl,
		minSize: 1024, // 1KB
		maxSize: 8192, // 8KB
	}
}

// ValueIterator 的方法
func (it *ValueIterator[T]) Next() bool {
	if it.current == nil {
		return false
	}

	it.pos++
	if int(it.pos) >= int(it.current.size-it.current.offset) {
		it.current = it.current.next
		it.pos = 0
	}
	return it.current != nil
}

func (it *ValueIterator[T]) V() T { // 改名：Current -> V
	var zero T
	if it.current == nil {
		return zero
	}
	ptr := unsafe.Add(it.current.data,
		uintptr(it.current.offset+it.pos)*unsafe.Sizeof(zero))
	return *(*T)(ptr)
}

// ChunkIterator 的方法
func (it *ChunkIterator[T]) Next() bool {
	if it.current == nil {
		return false
	}

	size := it.current.size
	offset := it.current.offset
	if size <= offset {
		it.current = it.current.next
		return it.Next()
	}

	validCount := size - offset
	if validCount < it.minSize && it.current.next != nil {
		it.current = it.current.next
		return it.Next()
	}

	if validCount > it.maxSize {
		validCount = it.maxSize
	}

	// 準備當前塊
	it.chunk = make([]T, validCount)
	src := unsafe.Add(it.current.data,
		uintptr(offset)*unsafe.Sizeof(it.chunk[0]))
	srcSlice := unsafe.Slice((*T)(src), validCount)
	copy(it.chunk, srcSlice)

	if validCount >= it.maxSize {
		it.current.offset += validCount
	} else {
		it.current = it.current.next
	}

	return true
}

func (it *ChunkIterator[T]) V() []T { // 改名：Current -> V
	return it.chunk
}
