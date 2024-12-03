package chunkpipe

// ChunkPipeNew 創建一個新的 ChunkPipe 實例
func ChunkPipeNew[T any]() *ChunkPipe[T] {
	return &ChunkPipe[T]{}
}
