package chunkpipe

import (
	"sync/atomic"
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

	dataLen := len(data)
	// 小數據快速路徑 (<=32 bytes)
	if dataLen <= 4 {
		cl.pushMu.Lock()
		defer cl.pushMu.Unlock()

		// 使用堆棧分配
		var stackData [4]T
		copy(stackData[:], data)

		block := &Chunk[T]{
			data:   unsafe.Pointer(&stackData[0]),
			size:   int32(dataLen),
			offset: 0,
		}

		if cl.tail == nil {
			cl.head = block
			cl.tail = block
		} else {
			block.prev = cl.tail
			cl.tail.next = block
			cl.tail = block
		}

		atomic.AddInt32(&cl.totalSize, int32(dataLen))
		atomic.AddInt32(&cl.validSize, int32(dataLen))
		return cl
	}

	// 大數據處理邏輯
	cl.pushMu.Lock()
	defer cl.pushMu.Unlock()

	// 分配新的數據塊
	block := &Chunk[T]{
		data:   unsafe.Pointer(&data[0]), // 直接使用輸入數據的指針
		size:   int32(dataLen),
		offset: 0,
	}

	// 更新鏈表
	if cl.tail == nil {
		cl.head = block
		cl.tail = block
	} else {
		block.prev = cl.tail
		cl.tail.next = block
		cl.tail = block
	}

	// 更新計數器
	atomic.AddInt32(&cl.totalSize, int32(dataLen))
	atomic.AddInt32(&cl.validSize, int32(dataLen))

	return cl
}

//go:nosplit
//go:noinline
func (cl *ChunkPipe[T]) Get(index int) (T, bool) {
	var zero T
	elemSize := unsafe.Sizeof(zero)

	// 快速路徑：檢查頭部（無需鎖）
	head := (*Chunk[T])(atomic.LoadPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&cl.head))))
	if head == nil {
		return zero, false
	}

	// 內聯所有計算
	offset := head.offset
	size := head.size
	validSize := size - offset

	// 快速路徑：頭部訪問
	if uint(index) < uint(validSize) {
		// 內聯指針計算
		ptr := unsafe.Add(head.data,
			uintptr(offset)*elemSize+uintptr(index)*elemSize)
		return *(*T)(ptr), true
	}

	// 檢查總大小
	if uint(index) >= uint(atomic.LoadInt32(&cl.validSize)) {
		return zero, false
	}

	// 慢路徑：需要遍歷
	cl.mu.RLock()
	current := head
	pos := int(validSize)

	for current = current.next; current != nil; current = current.next {
		offset = current.offset
		size = current.size
		blockSize := int(size - offset)
		nextPos := pos + blockSize

		if uint(index) < uint(nextPos) {
			// 內聯指針計算
			ptr := unsafe.Add(current.data,
				uintptr(offset)*elemSize+uintptr(index-pos)*elemSize)
			cl.mu.RUnlock()
			return *(*T)(ptr), true
		}
		pos = nextPos
	}

	cl.mu.RUnlock()
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

	// 創新的切片並安全複製數據
	newData := make([]T, validCount)
	if head.data != nil {
		src := unsafe.Slice((*T)(head.data), size)
		copy(newData, src[offset:])
	}

	// 更新表
	next := head.next
	cl.head = next
	if next != nil {
		next.prev = nil
	} else {
		cl.tail = nil
	}

	// 更新計數
	atomic.AddInt32(&cl.totalSize, -int32(validCount))
	atomic.AddInt32(&cl.validSize, -int32(validCount))

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
	atomic.AddInt32(&cl.totalSize, -int32(validCount))
	atomic.AddInt32(&cl.validSize, -int32(validCount))

	return newData, true
}

func (cl *ChunkPipe[T]) PopFront() (T, bool) {
	var zero T
	// 原子操作檢查大小
	if atomic.LoadInt32(&cl.validSize) == 0 {
		return zero, false
	}

	// 使用 CAS 獲取頭節點
	for {
		head := (*Chunk[T])(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cl.head))))
		if head == nil {
			return zero, false
		}

		// 快速路徑：使用原子操作
		offset := atomic.LoadInt32(&head.offset)
		if offset >= head.size {
			// 慢路徑：需要更新鏈表
			cl.mu.Lock()
			if head != cl.head {
				cl.mu.Unlock()
				continue
			}
			next := head.next
			if next != nil {
				next.prev = nil
				atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&cl.head)), unsafe.Pointer(next))
				// 預取
				_ = *(*byte)(next.data)
			} else {
				atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&cl.head)), nil)
				cl.tail = nil
			}
			cl.mu.Unlock()

			// 回收塊
			globalBlockCache.put((*Chunk[byte])(unsafe.Pointer(head)))
			return zero, false
		}

		// 嘗試子更新 offset
		if atomic.CompareAndSwapInt32(&head.offset, offset, offset+1) {
			// 讀取值
			ptr := unsafe.Add(head.data, uintptr(offset)*unsafe.Sizeof(zero))
			value := *(*T)(ptr)

			atomic.AddInt32(&cl.validSize, -1)
			atomic.AddInt32(&cl.totalSize, -1)
			return value, true
		}
	}
}

// // 抽取共用邏輯
// func (cl *ChunkPipe[T]) removeHead() {
// 	// 快取常用變數
// 	head := cl.head
// 	if head == nil {
// 		return
// 	}

// 	// 快取 next 指針
// 	next := head.next
// 	cl.head = next

// 	// 更新指針關係
// 	if next != nil {
// 		next.prev = nil
// 	} else {
// 		cl.tail = nil
// 	}

// 	// 清理原節點的指針
// 	head.next = nil
// 	head.prev = nil
// }

// // 新尾部移除方法
// func (cl *ChunkPipe[T]) removeTail() {
// 	// 快取常用變數
// 	tail := cl.tail
// 	if tail == nil {
// 		return
// 	}

// 	// 快取 prev 指針
// 	prev := tail.prev
// 	cl.tail = prev

// 	// 更新指針關係
// 	if prev != nil {
// 		prev.next = nil
// 	} else {
// 		cl.head = nil
// 	}

// 	// 清理原節點的指針
// 	tail.next = nil
// 	tail.prev = nil
// }

// 從尾部彈出數據
func (cl *ChunkPipe[T]) PopEnd() (T, bool) {
	var zero T
	cl.mu.RLock()
	tail := cl.tail
	if tail == nil || cl.validSize == 0 {
		cl.mu.RUnlock()
		return zero, false
	}
	cl.mu.RUnlock()

	// 快速路徑：使用原子操作
	size := atomic.LoadInt32(&tail.size)
	if size <= tail.offset {
		// 慢路徑：需要更新鏈表
		cl.mu.Lock()
		defer cl.mu.Unlock()

		prev := tail.prev
		if prev != nil {
			prev.next = nil
			cl.tail = prev
			// 預取前一個塊
			_ = *(*byte)(prev.data)
		} else {
			cl.head = nil
			cl.tail = nil
		}

		// 回收塊
		globalBlockCache.put((*Chunk[byte])(unsafe.Pointer(tail)))
		return zero, false
	}

	// 原子減少 size
	newSize := atomic.AddInt32(&tail.size, -1)
	atomic.AddInt32(&cl.validSize, -1)
	atomic.AddInt32(&cl.totalSize, -1)

	// 讀取值
	ptr := unsafe.Add(tail.data, uintptr(newSize)*unsafe.Sizeof(zero))
	value := *(*T)(ptr)

	return value, true
}

// ValueSlice 返回所有值的切片
func (cl *ChunkPipe[T]) ValueSlice() []T {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	// Check if the pipe is empty
	if cl.validSize == 0 {
		return []T{}
	}

	// Calculate total size needed
	totalSize := 0
	for current := cl.head; current != nil; current = current.next {
		totalSize += int(current.size - current.offset)
	}

	// Create result slice with exact capacity
	result := make([]T, 0, totalSize)

	// Safely append all values
	for current := cl.head; current != nil; current = current.next {
		if n := current.size - current.offset; n > 0 {
			src := unsafe.Slice((*T)(current.data), int(current.size))
			result = append(result, src[current.offset:current.size]...)
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
		minSize: 1024,
		maxSize: 8192,
	}
}

// ValueIterator 的方法
func (it *ValueIterator[T]) Next() bool {
	if it.current == nil {
		return false
	}

	// 先增加位置
	it.pos++

	// 如果超出當前塊的有效範圍，移動到下一個塊
	validSize := it.current.size - it.current.offset
	for int(it.pos) > int(validSize) {
		it.current = it.current.next
		if it.current == nil {
			return false
		}
		it.pos = 1 // 從新塊的第一個位置開始
		validSize = it.current.size - it.current.offset
	}

	return true
}

func (it *ValueIterator[T]) V() T {
	var zero T
	if it.current == nil {
		return zero
	}

	// 直接使用 pos-1 訪問當前位置的值
	ptr := unsafe.Add(it.current.data,
		uintptr(it.current.offset+it.pos-1)*unsafe.Sizeof(zero))
	return *(*T)(ptr)
}

// ChunkIterator 的方法
func (it *ChunkIterator[T]) Next() bool {
	if it.current == nil {
		return false
	}

	size := it.current.size
	offset := it.current.offset
	validCount := size - offset

	// 移除最小大小的檢查，讓所有塊都能被迭代
	if validCount <= 0 {
		it.current = it.current.next
		return it.Next()
	}

	// 準備當前塊
	it.chunk = make([]T, validCount)
	src := unsafe.Add(it.current.data,
		uintptr(offset)*unsafe.Sizeof(it.chunk[0]))
	srcSlice := unsafe.Slice((*T)(src), validCount)
	copy(it.chunk, srcSlice)

	// 移動到下一個塊
	it.current = it.current.next

	return true
}

func (it *ChunkIterator[T]) V() []T {
	return it.chunk
}
