package chunkpipe

import (
	"sync"
	"unsafe"
)

// 定義 Chunk 結構，用於存儲任意型別數據塊
type Chunk[T any] struct {
	data   unsafe.Pointer // 8 字節
	size   int32          // 4 字節
	offset int32          // 4 字節
	next   *Chunk[T]      // 8 字節
	prev   *Chunk[T]      // 8 字節
	_      [32]byte       // 32 字節的填充
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
	head      *Chunk[T]    // 頭節點
	tail      *Chunk[T]    // 尾節點
	totalSize int32        // 總大小
	validSize int32        // 有效大小
	mu        sync.RWMutex // 讀寫鎖
	pushMu    sync.Mutex   // Push 操作鎖
}

// 工廠函數：創建 ChunkPipe
func NewChunkPipe[T any]() *ChunkPipe[T] {
	return &ChunkPipe[T]{
		bptree:   NewBPTree[T](),
		skiplist: NewSkipList[T](),
		pool:     newMemoryPool(),
	}
}

// ValueIterator 提供值迭代器
type ValueIterator[T any] struct {
	current *Chunk[T]
	pos     int32
	pipe    *ChunkPipe[T]
}

// ChunkIterator 提供塊迭代器
type ChunkIterator[T any] struct {
	current *Chunk[T]
	pipe    *ChunkPipe[T]
	minSize int32 // 最小塊大小
	maxSize int32 // 最大塊大小
	chunk   []T   // 當前塊
}
