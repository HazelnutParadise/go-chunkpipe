//go:build !amd64 && !arm64

package chunkpipe

import (
	"unsafe"
	_ "unsafe"
)

func prefetch(addr unsafe.Pointer) {}
