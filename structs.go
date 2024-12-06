package chunkpipe

import "sync"

// 定義 Chunk 結構，用於存儲任意型別數據塊
type ChunkPipe[T any] struct {
	offset int
	list   []offset[T]
	mu     sync.RWMutex
}

type offset[T any] struct {
	off int
	val []T
}

func NewChunkPipe[T any]() *ChunkPipe[T] {
	return &ChunkPipe[T]{}
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
