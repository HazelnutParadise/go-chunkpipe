package chunkpipe

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"unsafe"
)

// 測試不同類型的數據結構
type TestStruct struct {
	ID   int
	Name string
	Data []byte
}

func TestDifferentTypes(t *testing.T) {
	// 測試整數
	t.Run("Integer", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		data := []int{1, 2, 3, 4, 5}
		cp.Push(data)

		if val, ok := cp.PopFront(); !ok || val != 1 {
			t.Errorf("Expected 1, got %v", val)
		}
	})

	// 測試字符串
	t.Run("String", func(t *testing.T) {
		cp := NewChunkPipe[string]()
		data := []string{"hello", "world"}
		cp.Push(data)

		if val, ok := cp.PopFront(); !ok || val != "hello" {
			t.Errorf("Expected 'hello', got %v", val)
		}
	})

	// 測試結構體
	t.Run("Struct", func(t *testing.T) {
		cp := NewChunkPipe[TestStruct]()
		data := []TestStruct{
			{ID: 1, Name: "test1", Data: []byte{1, 2, 3}},
			{ID: 2, Name: "test2", Data: []byte{4, 5, 6}},
		}
		cp.Push(data)

		if val, ok := cp.PopFront(); !ok || val.ID != 1 {
			t.Errorf("Expected ID 1, got %v", val.ID)
		}
	})
}

func TestConcurrency(t *testing.T) {
	cp := NewChunkPipe[int]()
	done := make(chan bool)

	// 並發寫入
	go func() {
		for i := 0; i < 1000; i++ {
			cp.Push([]int{i})
		}
		done <- true
	}()

	// 並發讀取
	go func() {
		count := 0
		iter := cp.ValueIter()
		for iter.Next() {
			_ = iter.V()
			count++
		}
		done <- true
	}()

	// 等待完成
	<-done
	<-done
}

func TestEdgeCases(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		cp := NewChunkPipe[byte]()
		if _, ok := cp.PopFront(); ok {
			t.Error("Should return false for empty pipe")
		}
	})

	t.Run("Large", func(t *testing.T) {
		cp := NewChunkPipe[byte]()
		data := make([]byte, 1000000)
		cp.Push(data)
		if cp.validSize != 1000000 {
			t.Errorf("Expected size 1000000, got %d", cp.validSize)
		}
	})

	t.Run("NilData", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		if _, ok := cp.PopFront(); ok {
			t.Error("PopFront should return false for nil data")
		}
		if _, ok := cp.PopEnd(); ok {
			t.Error("PopEnd should return false for nil data")
		}
	})

	t.Run("SingleElement", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		cp.Push([]int{1})
		if val, ok := cp.PopFront(); !ok || val != 1 {
			t.Errorf("PopFront failed: got %v, %v", val, ok)
		}
		if _, ok := cp.PopFront(); ok {
			t.Error("PopFront should return false after last element")
		}
	})

	t.Run("InvalidIndex", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		cp.Push([]int{1, 2, 3})
		if _, ok := cp.Get(-1); ok {
			t.Error("Get should return false for negative index")
		}
		if _, ok := cp.Get(3); ok {
			t.Error("Get should return false for out of range index")
		}
	})
}

func TestLargeDataHandling(t *testing.T) {
	sizes := []int{1 << 10, 1 << 15, 1 << 20} // 1KB, 32KB, 1MB

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size-%d", size), func(t *testing.T) {
			cp := NewChunkPipe[byte]()
			data := make([]byte, size)

			// 填充測試資料
			for i := range data {
				data[i] = byte(i % 256)
			}

			// 測試推送
			cp.Push(data)

			// 驗證大小
			if cp.validSize != int32(size) {
				t.Errorf("Expected size %d, got %d", size, cp.validSize)
			}

			// 測試讀取
			iter := cp.ValueIter()
			count := 0
			for iter.Next() {
				value := iter.V()
				if value != data[count] {
					t.Errorf("Data mismatch at index %d", count)
					break
				}
				count++
			}
			if count != int(size) {
				t.Errorf("Expected length %d, got %d", size, count)
			}
		})
	}
}

func TestIterators(t *testing.T) {
	pipe := NewChunkPipe[int]()
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	pipe.Push(data)

	t.Run("ValueIterator", func(t *testing.T) {
		iter := pipe.ValueIter()
		i := 0
		for iter.Next() {
			if got := iter.V(); got != data[i] {
				t.Errorf("value at %d = %v, want %v", i, got, data[i])
			}
			i++
		}
	})

	t.Run("ChunkIterator", func(t *testing.T) {
		iter := pipe.ChunkIter()
		total := 0
		for iter.Next() {
			chunk := iter.V()
			total += len(chunk)
			// 驗證塊內容
			for i, v := range chunk {
				if v != data[total-len(chunk)+i] {
					t.Errorf("chunk value at %d = %v, want %v",
						i, v, data[total-len(chunk)+i])
				}
			}
		}
		if total != len(data) {
			t.Errorf("total items = %d, want %d", total, len(data))
		}
	})
}

func TestPublicMethods(t *testing.T) {
	t.Run("Push", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		data := []int{1, 2, 3}
		cp.Push(data)
		if cp.validSize != 3 {
			t.Errorf("Push failed: expected size 3, got %d", cp.validSize)
		}
	})

	t.Run("Get", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		data := []int{1, 2, 3}
		cp.Push(data)
		if val, ok := cp.Get(1); !ok || val != 2 {
			t.Errorf("Get failed: expected 2, got %v", val)
		}
	})

	t.Run("PopFront", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		data := []int{1, 2, 3}
		cp.Push(data)
		if val, ok := cp.PopFront(); !ok || val != 1 {
			t.Errorf("PopFront failed: expected 1, got %v", val)
		}
	})

	t.Run("PopEnd", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		data := []int{1, 2, 3}
		cp.Push(data)
		if val, ok := cp.PopEnd(); !ok || val != 3 {
			t.Errorf("PopEnd failed: expected 3, got %v", val)
		}
	})

	t.Run("PopChunkFront", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		data := []int{1, 2, 3}
		cp.Push(data)
		if chunk, ok := cp.PopChunkFront(); !ok || len(chunk) != 3 || chunk[0] != 1 {
			t.Errorf("PopChunkFront failed: expected [1,2,3], got %v", chunk)
		}
	})

	t.Run("PopChunkEnd", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		data := []int{1, 2, 3}
		cp.Push(data)
		if chunk, ok := cp.PopChunkEnd(); !ok || len(chunk) != 3 || chunk[2] != 3 {
			t.Errorf("PopChunkEnd failed: expected [1,2,3], got %v", chunk)
		}
	})

	t.Run("ValueSlice", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		data := []int{1, 2, 3}
		cp.Push(data)
		slice := cp.ValueSlice()
		if len(slice) != 3 || slice[1] != 2 {
			t.Errorf("ValueSlice failed: expected [1,2,3], got %v", slice)
		}
	})

	t.Run("ChunkSlice", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		data := []int{1, 2, 3}
		cp.Push(data)
		chunks := cp.ChunkSlice()
		if len(chunks) != 1 || len(chunks[0]) != 3 || chunks[0][1] != 2 {
			t.Errorf("ChunkSlice failed: expected [[1,2,3]], got %v", chunks)
		}
	})
}

func TestMemoryPool(t *testing.T) {
	t.Run("Alloc", func(t *testing.T) {
		pool := newMemoryPool()
		ptr := pool.Alloc(1024)
		if ptr == nil {
			t.Error("Alloc failed")
		}
	})

	t.Run("Free", func(t *testing.T) {
		pool := newMemoryPool()

		// 先分配一些記憶體
		ptr := pool.Alloc(1024) // 分配 1KB

		beforeSize := pool.Size()
		pool.Free(ptr, 1024) // 添加必要的參數
		afterSize := pool.Size()

		if beforeSize == 0 {
			t.Error("初始大小不應為 0")
		}
		if afterSize != 0 {
			t.Errorf("Free 後大小應為 0，但得到 %d", afterSize)
		}
	})
}

func TestBlockCache(t *testing.T) {
	t.Run("PutGet", func(t *testing.T) {

		block := &Chunk[byte]{
			data:   unsafe.Pointer(&[1]byte{1}),
			size:   1,
			offset: 0,
		}

		globalBlockCache.put(block)
		got := globalBlockCache.get()
		if got == nil {
			t.Error("get returned nil")
		}
	})

	t.Run("EmptyGet", func(t *testing.T) {
		got := globalBlockCache.get()
		if got != nil {
			t.Error("get should return nil for empty cache")
		}
	})
}

func TestSIMD(t *testing.T) {
	t.Run("SmallCopy", func(t *testing.T) {
		src := []byte{1, 2, 3, 4}
		dst := make([]byte, 4)
		simdCopy(unsafe.Pointer(&dst[0]), unsafe.Pointer(&src[0]), 4)
		if !bytes.Equal(src, dst) {
			t.Errorf("copy failed: got %v, want %v", dst, src)
		}
	})

	t.Run("LargeCopy", func(t *testing.T) {
		src := make([]byte, 1024)
		for i := range src {
			src[i] = byte(i)
		}
		dst := make([]byte, 1024)
		simdCopy(unsafe.Pointer(&dst[0]), unsafe.Pointer(&src[0]), 1024)
		if !bytes.Equal(src, dst) {
			t.Error("large copy failed")
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	cp := NewChunkPipe[int]()
	const goroutines = 10
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // readers + writers

	// Writers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cp.Push([]int{id*iterations + j})
			}
		}(i)
	}

	// Readers
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cp.ValueIter()
				cp.ChunkIter()
				cp.ValueSlice()
				cp.ChunkSlice()
			}
		}()
	}

	wg.Wait()
}

func TestValueIteratorEdgeCases(t *testing.T) {
	t.Run("EmptyIterator", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		iter := cp.ValueIter()
		if iter.Next() {
			t.Error("Next should return false for empty iterator")
		}
		if iter.V() != 0 {
			t.Error("V should return zero value for empty iterator")
		}
	})
}

func TestChunkIteratorEdgeCases(t *testing.T) {
	t.Run("SmallChunks", func(t *testing.T) {
		cp := NewChunkPipe[int]()
		for i := 0; i < 100; i++ {
			cp.Push([]int{i})
		}
		iter := cp.ChunkIter()
		chunks := 0
		for iter.Next() {
			chunk := iter.V()
			if len(chunk) < 1 {
				t.Error("chunk size should be at least 1")
			}
			chunks++
		}
		if chunks == 0 {
			t.Error("should have at least one chunk")
		}
	})
}
