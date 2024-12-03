//go:build amd64 || arm64

package chunkpipe

import (
	"unsafe"
	_ "unsafe"
)

//go:linkname prefetch runtime.prefetch
func prefetch(addr unsafe.Pointer)
