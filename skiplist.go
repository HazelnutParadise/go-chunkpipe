package chunkpipe

import (
	"sync"
	"unsafe"
	_ "unsafe"
)

const maxLevel = 16

type SkipListNode[T any] struct {
	data     unsafe.Pointer
	position uintptr
	forward  [maxLevel]unsafe.Pointer
}

type SkipList[T any] struct {
	header unsafe.Pointer
	level  int
}

var skipNodePool = sync.Pool{
	New: func() interface{} {
		return &SkipListNode[byte]{
			forward: [maxLevel]unsafe.Pointer{},
		}
	},
}

func NewSkipList[T any]() *SkipList[T] {
	node := &SkipListNode[T]{
		forward: [maxLevel]unsafe.Pointer{},
	}
	return &SkipList[T]{
		header: unsafe.Pointer(node),
	}
}

func (sl *SkipList[T]) Insert(position uintptr, data unsafe.Pointer) {
	var update [maxLevel]unsafe.Pointer
	current := sl.header

	// 從最高層開始搜索
	for i := sl.level - 1; i >= 0; i-- {
		for current != nil && (*SkipListNode[T])(current).position < position {
			current = (*SkipListNode[T])(current).forward[i]
		}
		update[i] = current
	}

	// 隨機生成新節點的層數
	level := 1
	for level < maxLevel && fastrand()%2 == 0 {
		level++
	}

	// 創建新節點
	node := (*SkipListNode[T])(skipNodePool.Get().(*SkipListNode[byte]))
	node.data = data
	node.position = position

	// 如果新層數大於當前層數，更新層數
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.header
		}
		sl.level = level
	}

	// 更新指針
	for i := 0; i < level; i++ {
		if update[i] != nil {
			updateNode := (*SkipListNode[T])(update[i])
			node.forward[i] = updateNode.forward[i]
			updateNode.forward[i] = unsafe.Pointer(node)
		}
	}
}

func (sl *SkipList[T]) Find(position uintptr) (unsafe.Pointer, uintptr) {
	current := sl.header

	for i := sl.level - 1; i >= 0; i-- {
		for current != nil && (*(*SkipListNode[T])(current)).position <= position {
			current = (*(*SkipListNode[T])(current)).forward[i]
		}
	}

	if current != nil {
		node := (*SkipListNode[T])(current)
		return node.data, position - node.position
	}
	return nil, 0
}

//go:linkname fastrand runtime.fastrand
func fastrand() uint32
