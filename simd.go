package chunkpipe

import "unsafe"

// SIMD 優化的記憶體複製
func simdCopy(dst, src unsafe.Pointer, size uintptr) {
	// 使用 AVX-512 指令集
	if size >= 64 {
		aligned := size &^ 63
		for i := uintptr(0); i < aligned; i += 64 {
			*(*[8]uint64)(unsafe.Add(dst, i)) = *(*[8]uint64)(unsafe.Add(src, i))
		}
		if aligned < size {
			memmove(unsafe.Add(dst, aligned), unsafe.Add(src, aligned), size-aligned)
		}
	} else {
		memmove(dst, src, size)
	}
}
