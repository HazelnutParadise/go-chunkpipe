# ChunkPipe 塊聯管

塊聯管是一個高性能的 Go 管道資料結構，它將插入的每個切片視為一個塊，並將多個塊連結成一個管道。

塊聯管可以一次存入或取出單個元素或整個塊，非常適合作為整塊資料的儲存與操作（例如需要資料連續性的場景），也適合作為 Queue 或 Stack 使用。

## 系統要求

- Go 1.22.7 或更高版本
- 支援 x86-64 架構
- 支援 Linux/Windows/macOS

## 特點

- 零分配：大部分操作實現 0 allocs/op
- 高性能：比原生切片快 5-10 倍的迭代速度
- 泛型支援：可處理任意類型資料
- 天生併發安全：優化的原子操作和鎖機制
- SIMD 加速：使用 CPU 向量指令優化
- 記憶體效率：智慧記憶體池管理

## 安裝

```bash
go get -u github.com/HazelnutParadise/go-chunkpipe@latest
```

## 使用

```go
import (
	chunkpipe "github.com/HazelnutParadise/go-chunkpipe"
)
```

### 初始化

```go
cp := chunkpipe.NewChunkPipe[type]()
```

### 基礎操作

#### 插入

```go
cp.Push(data)
```

#### 取出

1. 取出第一個元素
```go
cp.PopFront()
```

2. 取出最後一個元素
```go
cp.PopEnd()
```

3. 取出第一個塊
```go
cp.PopChunkFront()
```

4. 取出最後一個塊
```go
cp.PopChunkEnd()
```

#### 隨機訪問

```go
cp.Get(index)
```

#### 迭代器

```go
iter := cp.ValueIter()
for iter.Next() {
    value := iter.V()
    // do something with value
}
```

## 性能

```bash
goos: darwin
goarch: amd64
pkg: github.com/HazelnutParadise/go-chunkpipe
cpu: Intel(R) Core(TM) i5-8257U CPU @ 1.40GHz
BenchmarkPush/ChunkPipe-10-8      	10852458	        99.44 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-10-8          	100000000	        47.17 ns/op	      57 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100-8     	11080180	       115.0 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-100-8         	19295209	      1465 ns/op	     580 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000-8    	 5969926	       201.7 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-1000-8        	  823974	     11018 ns/op	    5569 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-10000-8   	 8047227	       173.8 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-10000-8       	   74509	     34425 ns/op	   61589 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100000-8  	 8449328	       148.5 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-100000-8      	   10279	    898422 ns/op	  536126 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000000-8 	 5563766	       216.2 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-1000000-8     	     994	   8843507 ns/op	 5153452 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-_-8         	19951656	        55.40 ns/op	       6 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-_-8           	17047867	        71.66 ns/op	       6 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-_-8             	65517092	        18.48 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-_-8               	84942870	        13.66 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-_-8    	15581691	        73.53 ns/op	       8 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-_-8      	16808436	        71.34 ns/op	       8 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-_-8             	78184090	        15.21 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-d-8         	32352661	        37.45 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-d-8           	22295314	        53.56 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-d-8             	87428317	        13.50 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-d-8               	100000000	        11.46 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-d-8    	22352256	        51.24 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-d-8      	23952752	        50.60 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-d-8             	100000000	        11.29 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-Ϩ-8         	36838998	        31.47 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-Ϩ-8           	26734312	        45.73 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-Ϩ-8             	100000000	        11.57 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-Ϩ-8               	100000000	        10.34 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-Ϩ-8    	27150525	        45.03 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-Ϩ-8      	27243542	        44.63 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-Ϩ-8             	124500372	         9.591 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-✐-8         	41549622	        29.14 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-✐-8           	28699004	        41.83 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-✐-8             	100000000	        10.48 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-✐-8               	100000000	        10.33 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-✐-8    	29093899	        42.34 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-✐-8      	26028866	        41.70 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-✐-8             	127529554	         9.235 ns/op	       1 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8          	 2182626	       551.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8          	17107285	        63.58 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-100-8         	17092276	        70.28 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	 9910669	       128.5 ns/op	     136 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-100-8        	31222996	        47.48 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	28320963	        39.29 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	  224114	      5399 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	 3847790	       307.0 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	 3837314	       323.3 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	 3325112	       380.1 ns/op	    1048 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 3503433	       342.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 3494676	       343.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	   21046	     58528 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	 1761134	       679.7 ns/op	    2048 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  514585	      2314 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	  490382	      2511 ns/op	   10264 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  321018	      3739 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  322006	      3734 ns/op	       0 B/op	       0 allocs/op
BenchmarkMemoryOperations/Alloc-64-8        	 5316500	       214.3 ns/op	     120 B/op	       2 allocs/op
BenchmarkMemoryOperations/Alloc-1024-8      	 1759244	       584.5 ns/op	    1080 B/op	       2 allocs/op
BenchmarkMemoryOperations/Alloc-4096-8      	 1000000	      3337 ns/op	    4156 B/op	       2 allocs/op
BenchmarkConcurrentOperations-8             	 3420014	       367.6 ns/op	     120 B/op	       3 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	 5349925	       216.1 ns/op	     104 B/op	       3 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	   10000	   2989714 ns/op	 5120621 B/op	    5001 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	    1923	    668609 ns/op	 1048651 B/op	     129 allocs/op
BenchmarkGet/ChunkPipe-Get-100-8            	51408282	        23.37 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-100-8                	82648202	        14.51 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-1000-8           	50457261	        23.37 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-1000-8               	82895036	        14.72 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-10000-8          	51487338	        23.53 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-10000-8              	83044028	        14.52 ns/op	       0 B/op	       0 allocs/op
```