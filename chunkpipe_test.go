package chunkpipe

import (
	"fmt"
	"testing"
)

// 基準測試：插入操作
func BenchmarkPush(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000, 1000000}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		b.Run(fmt.Sprintf("ChunkPipe-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cp.Push(data)
			}
		})

		b.Run(fmt.Sprintf("Slice-%d", size), func(b *testing.B) {
			var slice []byte
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				slice = append(slice, data...)
			}
		})
	}
}

// 基準測試：彈出操作
func BenchmarkPop(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		b.Run("ChunkPipe-PopFront-"+string(rune(size)), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			cp.Push(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i%size == 0 {
					cp.Push(data)
				}
				cp.PopFront()
			}
		})

		b.Run("Slice-PopFront-"+string(rune(size)), func(b *testing.B) {
			slice := make([]byte, size)
			copy(slice, data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i%size == 0 {
					slice = append(slice, data...)
				}
				if len(slice) > 0 {
					slice = slice[1:]
				}
			}
		})
	}
}

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

// 測試並發安全性
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

// 測試極限情況
func TestEdgeCases(t *testing.T) {
	cp := NewChunkPipe[byte]()

	// 測試空數據
	t.Run("Empty", func(t *testing.T) {
		if _, ok := cp.PopFront(); ok {
			t.Error("Should return false for empty pipe")
		}
	})

	// 測試大量數據
	t.Run("Large", func(t *testing.T) {
		data := make([]byte, 1000000)
		cp.Push(data)
		if cp.validSize != 1000000 {
			t.Errorf("Expected size 1000000, got %d", cp.validSize)
		}
	})
}

// 測試內存使用
func BenchmarkMemoryUsage(b *testing.B) {
	sizes := []int{1024, 1024 * 1024}
	for _, size := range sizes {
		data := make([]byte, size)

		b.Run(fmt.Sprintf("ChunkPipe-%d", size), func(b *testing.B) {
			b.ReportAllocs()
			cp := NewChunkPipe[byte]()
			for i := 0; i < b.N; i++ {
				cp.Push(data)
				iter := cp.ChunkIter()
				for iter.Next() {
					_ = iter.V()
				}
			}
		})
	}
}

// 測試混合操作
func BenchmarkMixedOperations(b *testing.B) {
	b.Run("ChunkPipe", func(b *testing.B) {
		cp := NewChunkPipe[int]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cp.Push([]int{i})
			if i%2 == 0 {
				cp.PopFront()
			} else {
				cp.PopEnd()
			}
		}
	})
}

// 增加新的測試案例
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
			if cp.validSize != size {
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
			if count != size {
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

// 基準測試：迭代器操作
func BenchmarkIterators(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		b.Run(fmt.Sprintf("ValueIter-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			cp.Push(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				iter := cp.ValueIter()
				for iter.Next() {
					_ = iter.V()
				}
			}
		})

		b.Run(fmt.Sprintf("ChunkIter-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			cp.Push(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				iter := cp.ChunkIter()
				for iter.Next() {
					_ = iter.V()
				}
			}
		})

		b.Run(fmt.Sprintf("ValueSlice-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			cp.Push(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = cp.ValueSlice()
			}
		})

		b.Run(fmt.Sprintf("ChunkSlice-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			cp.Push(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = cp.ChunkSlice()
			}
		})
	}
}

// 測試所有公開方法
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
