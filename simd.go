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
