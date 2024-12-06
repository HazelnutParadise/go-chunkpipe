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

		b.Run(fmt.Sprintf("ChunkPipe-PopFront-%d", size), func(b *testing.B) {
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

		b.Run(fmt.Sprintf("ChunkPipe-PopEnd-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			cp.Push(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i%size == 0 {
					cp.Push(data)
				}
				cp.PopEnd()
			}
		})

		b.Run(fmt.Sprintf("Slice-PopFront-%d", size), func(b *testing.B) {
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

		b.Run(fmt.Sprintf("Slice-PopEnd-%d", size), func(b *testing.B) {
			slice := make([]byte, size)
			copy(slice, data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i%size == 0 {
					slice = append(slice, data...)
				}
				if len(slice) > 0 {
					slice = slice[:len(slice)-1]
				}
			}
		})

		b.Run(fmt.Sprintf("ChunkPipe-PopChunkFront-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			cp.Push(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i%size == 0 {
					cp.Push(data)
				}
				cp.PopChunkFront()
			}
		})

		b.Run(fmt.Sprintf("ChunkPipe-PopChunkEnd-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			cp.Push(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i%size == 0 {
					cp.Push(data)
				}
				cp.PopChunkEnd()
			}
		})

		b.Run(fmt.Sprintf("Slice-PopChunk-%d", size), func(b *testing.B) {
			slice := make([]byte, size)
			copy(slice, data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if i%size == 0 {
					slice = append(slice, data...)
				}
				if len(slice) > 0 {
					chunk := make([]byte, len(slice))
					copy(chunk, slice)
					slice = slice[:0]
				}
			}
		})
	}
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

		b.Run(fmt.Sprintf("NativeSlice-%d", size), func(b *testing.B) {
			slice := make([]byte, size)
			copy(slice, data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for range slice {
				}
			}
		})

		b.Run(fmt.Sprintf("NativeSliceValue-%d", size), func(b *testing.B) {
			slice := make([]byte, size)
			copy(slice, data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, v := range slice {
					_ = v
				}
			}
		})
	}
}

// 基準測試：內存操作
func BenchmarkMemoryOperations(b *testing.B) {
	sizes := []int{64, 1024, 4096}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Alloc-%d", size), func(b *testing.B) {
			pool := newMemoryPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pool.Alloc(uintptr(size))
			}
		})
	}
}

// 基準測試：並發操作
func BenchmarkConcurrentOperations(b *testing.B) {
	cp := NewChunkPipe[int]()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cp.Push([]int{1, 2, 3})
			cp.PopFront()
		}
	})
}

// 基準測試：混合操作
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

// 基準測試：內存使用
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

// 基準測試：隨機訪問操作
func BenchmarkGet(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		b.Run(fmt.Sprintf("ChunkPipe-Get-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			cp.Push(data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				index := i % size
				_, _ = cp.Get(index)
			}
		})

		b.Run(fmt.Sprintf("Slice-Get-%d", size), func(b *testing.B) {
			slice := make([]byte, size)
			copy(slice, data)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				index := i % size
				_ = slice[index]
			}
		})
	}
}

// 基準測試：批量推送操作
func BenchmarkPushBatch(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		b.Run(fmt.Sprintf("ChunkPipe-Push-%d", size), func(b *testing.B) {
			cp := NewChunkPipe[byte]()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cp.Push(data)
				if i%100 == 99 {
					cp = NewChunkPipe[byte]()
				}
			}
		})

		b.Run(fmt.Sprintf("Slice-Append-%d", size), func(b *testing.B) {
			var slice []byte
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				slice = append(slice, data...)
				if i%100 == 99 {
					slice = nil // 定期重置
				}
			}
		})
	}
}
