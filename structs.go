package chunkpipe

import (
	"runtime"
	"sync"
)

type ChunkPipe[T any] struct {
	mu       sync.RWMutex
	listPool sync.Pool
	offset   int
	// 新增 pools
	valueSlicePool sync.Pool
	chunkSlicePool sync.Pool
}

type offset[T any] struct {
	off int
	val []T
}

// 在 ChunkPipe 結構體中修改 New 函數的返回類型
func NewChunkPipe[T any]() *ChunkPipe[T] {
	cp := &ChunkPipe[T]{
		listPool: sync.Pool{
			New: func() interface{} {
				return &[]offset[T]{}
			},
		},
		chunkSlicePool: sync.Pool{
			New: func() interface{} {
				slice := make([][]T, 1024)
				return &slice // 返回指針
			},
		},
		valueSlicePool: sync.Pool{
			New: func() interface{} {
				slice := make([]T, 1024)
				return &slice // 返回指針
			},
		},
	}

	go func() {
		runtime.KeepAlive(&cp.listPool)
		runtime.KeepAlive(&cp.valueSlicePool)
		runtime.KeepAlive(&cp.chunkSlicePool)
	}()
	return cp
}

// ValueIterator 提供值迭代器
type ValueIterator[T any] struct {
	pos  int
	pipe *ChunkPipe[T]
}

// ChunkIterator 提供塊迭代器
type ChunkIterator[T any] struct {
	pos  int
	pipe *ChunkPipe[T]
}
