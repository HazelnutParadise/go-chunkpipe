# ChunkPipe 塊聯管

塊聯管是一個高性能的 Go 管道資料結構，它將插入的每個切片視為一個塊，並將多個塊連結成一個管道。

塊聯管可以一次存入或取出單個元素或整個塊，非常適合作為整塊資料的儲存與操作（例如需要資料連續性的場景），也適合作為 Queue 或 Stack 使用。

## 系統要求

- Go 1.22.7 或更高版本
- 支援 x86-64 架構
- 支援 Linux/Windows/macOS

## 特點

- 零分配：所有操作實現 0 allocs/op
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
import "github.com/HazelnutParadise/go-chunkpipe"
```

### 初始化

```go
cp := chunkpipe.NewChunkPipe[int]()
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
BenchmarkPush/ChunkPipe-10-8      	 6469540	       192.2 ns/op	     128 B/op	       1 allocs/op
BenchmarkPush/Slice-10-8          	86582197	        41.73 ns/op	      53 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100-8     	10204665	       125.5 ns/op	     128 B/op	       1 allocs/op
BenchmarkPush/Slice-100-8         	 9019784	       282.2 ns/op	     508 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000-8    	 8613740	       128.6 ns/op	     128 B/op	       1 allocs/op
BenchmarkPush/Slice-1000-8        	 1000000	      9056 ns/op	    5736 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-10000-8   	 2458242	       568.3 ns/op	     128 B/op	       1 allocs/op
BenchmarkPush/Slice-10000-8       	   30960	     79017 ns/op	   60693 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100000-8  	 7226866	       282.5 ns/op	     128 B/op	       1 allocs/op
BenchmarkPush/Slice-100000-8      	   10000	   1520979 ns/op	  551084 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000000-8 	 2287266	       566.2 ns/op	     128 B/op	       1 allocs/op
BenchmarkPush/Slice-1000000-8     	     790	  13402391 ns/op	 5184012 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-_-8         	14950482	        80.58 ns/op	      12 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-_-8           	12061034	        98.74 ns/op	      12 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-_-8             	45473422	        25.19 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-_-8               	62146810	        19.36 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-_-8    	11193272	       107.9 ns/op	      14 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-_-8      	11198365	       108.1 ns/op	      14 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-_-8             	52407682	        23.21 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-d-8         	20495497	        58.97 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-d-8           	14473975	        80.92 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-d-8             	60124387	        20.25 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-d-8               	66500323	        17.08 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-d-8    	15082972	        76.49 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-d-8      	15789262	        76.20 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-d-8             	69515887	        16.75 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-Ϩ-8         	25220313	        47.15 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-Ϩ-8           	17498684	        69.31 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-Ϩ-8             	63959406	        17.45 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-Ϩ-8               	80439952	        14.92 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-Ϩ-8    	18090754	        65.01 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-Ϩ-8      	18759286	        63.13 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-Ϩ-8             	85082095	        13.75 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-✐-8         	29412513	        40.82 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-✐-8           	19278744	        59.29 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-✐-8             	76043409	        15.14 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-✐-8               	89024680	        13.60 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-✐-8    	20797171	        57.55 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-✐-8      	21014379	        58.48 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-✐-8             	93402200	        12.35 ns/op	       1 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8          	 1534708	       737.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8          	13656150	        84.03 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-100-8         	12387804	        94.05 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	 7774488	       154.5 ns/op	     136 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-100-8        	21208251	        54.05 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	21064732	        55.73 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	  174892	      6872 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	 3659076	       337.4 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	 2979154	       418.5 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	 2975215	       398.9 ns/op	    1048 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 2905004	       415.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 2910250	       411.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	   17534	     68571 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	 1810770	       663.9 ns/op	    2048 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  386854	      3173 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	  510951	      2428 ns/op	   10264 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  292321	      4036 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  298208	      4033 ns/op	       0 B/op	       0 allocs/op
BenchmarkMemoryOperations/Alloc-64-8        	 3370107	       361.7 ns/op	     209 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-1024-8      	 2197478	       599.5 ns/op	    1166 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-4096-8      	 1000000	      2565 ns/op	    4223 B/op	       1 allocs/op
BenchmarkConcurrentOperations-8             	 3064042	       407.3 ns/op	     184 B/op	       3 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	 4506408	       269.3 ns/op	     168 B/op	       3 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	   10000	   3031140 ns/op	 5120706 B/op	    5002 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	    2100	    642792 ns/op	 1048714 B/op	     129 allocs/op
BenchmarkGet/ChunkPipe-Get-100-8            	48734220	        25.53 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-100-8                	78490614	        17.17 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-1000-8           	40560032	        29.39 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-1000-8               	78709308	        15.34 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-10000-8          	49160175	        25.02 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-10000-8              	78374215	        15.29 ns/op	       0 B/op	       0 allocs/op
```