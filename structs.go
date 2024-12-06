package chunkpipe

import "sync"

type ChunkPipe[T any] struct {
	mu     sync.RWMutex
	list   []offset[T]
	offset int
	// 新增 pools
	valuePool sync.Pool
	chunkPool sync.Pool
}

type offset[T any] struct {
	off int
	val []T
}

// 修改 NewChunkPipe 函數來初始化 pools
func NewChunkPipe[T any]() *ChunkPipe[T] {
	cp := &ChunkPipe[T]{
		valuePool: sync.Pool{
			New: func() interface{} {
				return make([]T, 0)
			},
		},
		chunkPool: sync.Pool{
			New: func() interface{} {
				return make([][]T, 0)
			},
		},
	}
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
