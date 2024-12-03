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

	baseLevel := uintptr(sl.level - 1)
	currentNode := (*SkipListNode[T])(current)

	levelSize := unsafe.Sizeof(unsafe.Pointer(nil))
	for level := baseLevel; ; level-- {
		forwardPtr := unsafe.Add(unsafe.Pointer(&currentNode.forward[0]), level*levelSize)
		nextPtr := *(*unsafe.Pointer)(forwardPtr)

		for nextPtr != nil {
			nextNode := (*SkipListNode[T])(nextPtr)
			if nextNode.position >= position {
				break
			}
			current = nextPtr
			currentNode = nextNode
			nextPtr = *(*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(&nextNode.forward[0]), level*levelSize))
		}

		update[level] = current
		if level == 0 {
			break
		}
	}

	node := (*SkipListNode[T])(skipNodePool.Get().(*SkipListNode[byte]))
	node.data = data
	node.position = position

	newLevel := uintptr(1)
	for newLevel < maxLevel && fastrand()&1 == 0 {
		newLevel++
	}

	if newLevel > uintptr(sl.level) {
		headerPtr := unsafe.Pointer(&sl.header)
		for i := uintptr(sl.level); i < newLevel; i++ {
			update[i] = headerPtr
		}
		sl.level = int(newLevel)
	}

	nodeForwardBase := unsafe.Pointer(&node.forward[0])
	for i := uintptr(0); i < newLevel; i++ {
		updateNode := (*SkipListNode[T])(update[i])
		updateForwardPtr := unsafe.Add(unsafe.Pointer(&updateNode.forward[0]), i*levelSize)

		*(*unsafe.Pointer)(unsafe.Add(nodeForwardBase, i*levelSize)) =
			*(*unsafe.Pointer)(updateForwardPtr)

		*(*unsafe.Pointer)(updateForwardPtr) = unsafe.Pointer(node)
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
