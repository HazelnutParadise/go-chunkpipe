package chunkpipe

import (
	"sync"
	"unsafe"
)

// 定義 Chunk 結構，用於存儲任意型別數據塊
type Chunk[T any] struct {
	data   unsafe.Pointer // 指向數據的指針
	size   int            // 數據大小
	offset int            // 當前讀取位置
	next   *Chunk[T]      // 下一個塊的指標
	prev   *Chunk[T]      // 前一個塊的指標
}

// 定義 TreeNode 結構，用於索引
type TreeNode[T any] struct {
	sum       int            // 當前節點及子節點的數據總大小
	validSize int            // 實際可用的數據大小（考慮offset後）
	blockAddr unsafe.Pointer // 指向數據塊的指針
	left      *TreeNode[T]   // 左子節點
	right     *TreeNode[T]   // 右子節點
}

// 主結構 ChunkPipe
type ChunkPipe[T any] struct {
	root      *TreeNode[T] // 原有的樹
	bptree    *BPTree[T]   // B+ Tree 索引
	skiplist  *SkipList[T] // Skip List 索引
	pool      *MemoryPool  // 記憶體池
	head      *Chunk[T]
	tail      *Chunk[T]
	totalSize int
	validSize int
	mu        sync.RWMutex
}

// 工廠函數：創建 ChunkPipe
func NewChunkPipe[T any]() *ChunkPipe[T] {
	return &ChunkPipe[T]{
		bptree:   NewBPTree[T](),
		skiplist: NewSkipList[T](),
		pool:     NewMemoryPool(),
	}
}

// 對齊到 CPU Cache line
type alignedChunk[T any] struct {
	data   unsafe.Pointer
	size   int32
	offset int32
	_      [56]byte // 填充到 64 字節
}
