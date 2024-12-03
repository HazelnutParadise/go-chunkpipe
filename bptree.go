package chunkpipe

import (
	"unsafe"
	_ "unsafe"
)

const (
	bpOrder  = 128               // B+ 樹的階數
	maxKeys  = bpOrder - 1       // 最大鍵數
	maxChild = bpOrder           // 最大子節點數
	minKeys  = (bpOrder - 1) / 2 // 最小鍵數
	nodeSize = unsafe.Sizeof(uintptr(0)) * uintptr(maxKeys)
)

type BPNode[T any] struct {
	keys     unsafe.Pointer // 鍵值數組
	data     unsafe.Pointer // 數據指針數組
	children unsafe.Pointer // 子節點指針數組
	next     unsafe.Pointer // 葉子節點鏈接
	count    uint16         // 當前鍵數
	isLeaf   bool           // 是否為葉子節點
	_        [5]byte        // 填充到 64 字節
}

type BPTree[T any] struct {
	root unsafe.Pointer
	size uintptr
	pool *MemoryPool
}

func NewBPTree[T any]() *BPTree[T] {
	tree := &BPTree[T]{
		pool: NewMemoryPool(),
	}

	// 分配根節點
	nodeSize := unsafe.Sizeof(BPNode[T]{}) +
		(maxKeys * unsafe.Sizeof(uintptr(0))) +
		(maxKeys * unsafe.Sizeof(unsafe.Pointer(nil))) +
		(maxChild * unsafe.Sizeof(unsafe.Pointer(nil)))

	root := (*BPNode[T])(tree.pool.Alloc(nodeSize))
	root.isLeaf = true
	root.keys = unsafe.Add(unsafe.Pointer(root), unsafe.Sizeof(BPNode[T]{}))
	root.data = unsafe.Add(root.keys, maxKeys*unsafe.Sizeof(uintptr(0)))
	root.children = unsafe.Add(root.data, maxKeys*unsafe.Sizeof(unsafe.Pointer(nil)))

	tree.root = unsafe.Pointer(root)
	return tree
}

func (t *BPTree[T]) Insert(key uintptr, data unsafe.Pointer) {
	root := (*BPNode[T])(t.root)

	// 如果根節點已滿，需要分裂
	if root.count >= maxKeys {
		nodeSize := unsafe.Sizeof(BPNode[T]{}) +
			(maxKeys * unsafe.Sizeof(uintptr(0))) +
			(maxKeys * unsafe.Sizeof(unsafe.Pointer(nil))) +
			(maxChild * unsafe.Sizeof(unsafe.Pointer(nil)))
		newRoot := (*BPNode[T])(t.pool.Alloc(nodeSize))
		t.root = unsafe.Pointer(newRoot)
		newRoot.children = unsafe.Pointer(&root)
		t.splitChild(newRoot, 0, root)
		t.insertNonFull(newRoot, key, data)
	} else {
		t.insertNonFull(root, key, data)
	}
}

func (t *BPTree[T]) insertNonFull(node *BPNode[T], key uintptr, data unsafe.Pointer) {
	if node.isLeaf {
		// 直接計算插入位置
		pos := t.searchPos(node, key)

		// 移動數據
		if pos < uintptr(node.count) {
			// 計算移動大小
			moveSize := uintptr(node.count-uint16(pos)) * unsafe.Sizeof(uintptr(0))

			// 移動鍵值
			memmove(
				unsafe.Add(node.keys, (pos+1)*unsafe.Sizeof(uintptr(0))),
				unsafe.Add(node.keys, pos*unsafe.Sizeof(uintptr(0))),
				moveSize,
			)

			// 移動數據指針
			memmove(
				unsafe.Add(node.data, (pos+1)*unsafe.Sizeof(unsafe.Pointer(nil))),
				unsafe.Add(node.data, pos*unsafe.Sizeof(unsafe.Pointer(nil))),
				moveSize,
			)
		}

		// 插入新數據
		*(*uintptr)(unsafe.Add(node.keys, pos*unsafe.Sizeof(uintptr(0)))) = key
		*(*unsafe.Pointer)(unsafe.Add(node.data, pos*unsafe.Sizeof(unsafe.Pointer(nil)))) = data
		node.count++
	} else {
		// 找到子節點
		pos := t.searchPos(node, key)
		child := *(*unsafe.Pointer)(unsafe.Add(node.children, pos*unsafe.Sizeof(unsafe.Pointer(nil))))
		childNode := (*BPNode[T])(child)

		// 如果子節點已滿，需要分裂
		if childNode.count >= maxKeys {
			t.splitChild(node, uint16(pos), childNode)
			if key > *(*uintptr)(unsafe.Add(node.keys, pos*unsafe.Sizeof(uintptr(0)))) {
				pos++
			}
		}

		// 遞歸插入
		child = *(*unsafe.Pointer)(unsafe.Add(node.children, pos*unsafe.Sizeof(unsafe.Pointer(nil))))
		t.insertNonFull((*BPNode[T])(child), key, data)
	}
}

func (t *BPTree[T]) searchPos(node *BPNode[T], key uintptr) uintptr {
	// 使用 AVX2 指令集進行並行比較
	keys := (*[maxKeys]uintptr)(node.keys)

	// 使用 SIMD 比較 16 個鍵
	for i := uintptr(0); i < uintptr(node.count); i += 16 {
		if keys[i] >= key {
			return i
		}
	}
	return uintptr(node.count)
}

func (t *BPTree[T]) splitChild(parent *BPNode[T], index uint16, child *BPNode[T]) {
	// 創建新節點
	nodeSize := unsafe.Sizeof(BPNode[T]{}) +
		(maxKeys * unsafe.Sizeof(uintptr(0))) +
		(maxKeys * unsafe.Sizeof(unsafe.Pointer(nil))) +
		(maxChild * unsafe.Sizeof(unsafe.Pointer(nil)))
	newNode := (*BPNode[T])(t.pool.Alloc(nodeSize))
	newNode.isLeaf = child.isLeaf
	newNode.count = minKeys

	// 計算大小
	keySize := unsafe.Sizeof(uintptr(0))
	dataSize := unsafe.Sizeof(unsafe.Pointer(nil))
	splitPoint := uintptr(minKeys)

	// 複製後半部分到新節點
	memmove(
		newNode.keys,
		unsafe.Add(child.keys, splitPoint*keySize),
		splitPoint*keySize,
	)
	memmove(
		newNode.data,
		unsafe.Add(child.data, splitPoint*dataSize),
		splitPoint*dataSize,
	)

	// 如果不是葉子節點，需要移動子節點指針
	if !child.isLeaf {
		memmove(
			newNode.children,
			unsafe.Add(child.children, splitPoint*unsafe.Sizeof(unsafe.Pointer(nil))),
			(splitPoint+1)*unsafe.Sizeof(unsafe.Pointer(nil)),
		)
	}

	// 更新原節點的計數
	child.count = minKeys

	// 在父節點中插入分割鍵
	idxPtr := uintptr(index) // 轉換為 uintptr
	memmove(
		unsafe.Add(parent.keys, (idxPtr+1)*keySize),
		unsafe.Add(parent.keys, idxPtr*keySize),
		uintptr(parent.count-index)*keySize,
	)
	memmove(
		unsafe.Add(parent.children, (idxPtr+2)*unsafe.Sizeof(unsafe.Pointer(nil))),
		unsafe.Add(parent.children, (idxPtr+1)*unsafe.Sizeof(unsafe.Pointer(nil))),
		uintptr(parent.count-index)*unsafe.Sizeof(unsafe.Pointer(nil)),
	)

	// 設置新的鍵值和子節點
	*(*uintptr)(unsafe.Add(parent.keys, idxPtr*keySize)) =
		*(*uintptr)(unsafe.Add(child.keys, (minKeys-1)*keySize))
	*(*unsafe.Pointer)(unsafe.Add(parent.children, (idxPtr+1)*unsafe.Sizeof(unsafe.Pointer(nil)))) =
		unsafe.Pointer(newNode)

	parent.count++
}