package chunkpipe

import (
	"fmt"
	"testing"
)

func benchmarkPush(b *testing.B, n, m int) {
	data := make([][]byte, m)
	for i := range data {
		data[i] = make([]byte, n)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cp := NewChunkPipe[byte]()
		for j := 0; j < m; j++ {
			cp.Push(data[j])
		}
	}
}

func BenchmarkPush10x10000(b *testing.B) {
	benchmarkPush(b, 10, 10000)
}

func BenchmarkPush100x1000(b *testing.B) {
	benchmarkPush(b, 100, 1000)
}

func BenchmarkPush1000x100(b *testing.B) {
	benchmarkPush(b, 1000, 100)
}

func BenchmarkPush10000x10(b *testing.B) {
	benchmarkPush(b, 10000, 10)
}

func generateData(n, m int) *ChunkPipe[byte] {
	data := make([][]byte, m)
	for i := range data {
		data[i] = make([]byte, n)
	}

	cp := NewChunkPipe[byte]()
	for j := 0; j < m; j++ {
		cp.Push(data[j])
	}
	return cp
}

func benchmarkPopEnd(b *testing.B, n int, m int) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cp := generateData(n, m)
		b.StartTimer()
		for j := 0; j < n*m; j++ {
			cp.PopEnd()
		}
	}
}

func benchmarkPopFront(b *testing.B, n int, m int) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cp := generateData(n, m)
		b.StartTimer()
		for j := 0; j < n*m; j++ {
			cp.PopFront()
		}
	}
}

func benchmarkPopChunkEnd(b *testing.B, n int, m int) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cp := generateData(n, m)
		b.StartTimer()
		for j := 0; j < m; j++ {
			cp.PopChunkEnd()
		}
	}
}

func benchmarkPopChunkFront(b *testing.B, n int, m int) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cp := generateData(n, m)
		b.StartTimer()
		for j := 0; j < m; j++ {
			cp.PopChunkFront()
		}
	}
}

func benchmarkGet(b *testing.B, n int, m int) {
	cp := generateData(n, m)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < n*m; j++ {
			cp.Get(j)
		}
	}
}

func BenchmarkPopEnd10x10000(b *testing.B) {
	benchmarkPopEnd(b, 10, 10000)
}

func BenchmarkPopEnd100x1000(b *testing.B) {
	benchmarkPopEnd(b, 100, 1000)
}

func BenchmarkPopEnd1000x100(b *testing.B) {
	benchmarkPopEnd(b, 1000, 100)
}

func BenchmarkPopEnd10000x10(b *testing.B) {
	benchmarkPopEnd(b, 10000, 10)
}

func BenchmarkPopChunkEnd10x10000(b *testing.B) {
	benchmarkPopChunkEnd(b, 10, 10000)
}

func BenchmarkPopChunkEnd100x1000(b *testing.B) {
	benchmarkPopChunkEnd(b, 100, 1000)
}

func BenchmarkPopChunkEnd1000x100(b *testing.B) {
	benchmarkPopChunkEnd(b, 1000, 100)
}

func BenchmarkPopFront10x10000(b *testing.B) {
	benchmarkPopFront(b, 10, 10000)
}

func BenchmarkPopFront100x1000(b *testing.B) {
	benchmarkPopFront(b, 100, 1000)
}

func BenchmarkPopFront1000x100(b *testing.B) {
	benchmarkPopFront(b, 1000, 100)
}

func BenchmarkPopFront10000x10(b *testing.B) {
	benchmarkPopFront(b, 10000, 10)
}

func BenchmarkPopChunkFront10x10000(b *testing.B) {
	benchmarkPopChunkFront(b, 10, 10000)
}

func BenchmarkPopChunkFront100x1000(b *testing.B) {
	benchmarkPopChunkFront(b, 100, 1000)
}

func BenchmarkPopChunkFront1000x100(b *testing.B) {
	benchmarkPopChunkFront(b, 1000, 100)
}

func BenchmarkGet10x10000(b *testing.B) {
	benchmarkGet(b, 10, 10000)
}

func BenchmarkGet100x1000(b *testing.B) {
	benchmarkGet(b, 100, 1000)
}

func BenchmarkGet1000x100(b *testing.B) {
	benchmarkGet(b, 1000, 100)
}

func BenchmarkGet10000x10(b *testing.B) {
	benchmarkGet(b, 10000, 10)
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
