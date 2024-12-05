package chunkpipe

import "unsafe"

//go:noescape
//go:linkname asmCopy runtime.memmove
func asmCopy(dst, src unsafe.Pointer, size uintptr)

func simdCopy(dst, src unsafe.Pointer, size uintptr) {
	if size >= 64 {
		// 使用匯編優化的複製
		asmCopy(dst, src, size)
	} else {
		// 小數據直接複製
		memmove(dst, src, size)
	}
}

// 使用純 Go 實現塊大小計算
func simdBlockSize(size, offset int32) int32 {
	return size - offset
}

// 使用純 Go 實現指針偏移計算
func simdPtrOffset(base unsafe.Pointer, offset, elemSize uintptr) unsafe.Pointer {
	return unsafe.Add(base, offset*elemSize)
}
