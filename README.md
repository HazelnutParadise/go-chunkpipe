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
BenchmarkPush/ChunkPipe-10-8      	11893615	        95.17 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-10-8          	100000000	        30.25 ns/op	      57 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100-8     	13946600	       115.8 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-100-8         	18662077	      1020 ns/op	     600 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000-8    	10125560	       124.5 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-1000-8        	 1000000	      9689 ns/op	    5736 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-10000-8   	 9925021	       103.2 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-10000-8       	  193929	     85653 ns/op	   57778 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100000-8  	11033799	       128.4 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-100000-8      	   10000	    659563 ns/op	  551084 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000000-8 	14641614	        72.24 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-1000000-8     	    2126	   4674679 ns/op	 5891527 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-_-8         	33954663	        34.06 ns/op	       6 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-_-8           	28815535	        45.58 ns/op	       6 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-_-8             	100000000	        10.98 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-_-8               	145319077	         8.370 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-_-8    	26152420	        45.59 ns/op	       8 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-_-8      	26528748	        45.16 ns/op	       8 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-_-8             	100000000	        10.06 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-d-8         	47881496	        25.52 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-d-8           	33822280	        35.49 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-d-8             	130250306	         9.223 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-d-8               	150397593	         8.065 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-d-8    	33120729	        36.50 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-d-8      	33874730	        36.76 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-d-8             	144720666	         8.264 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-Ϩ-8         	49506330	        23.91 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-Ϩ-8           	31981018	        34.89 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-Ϩ-8             	133236404	         9.033 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-Ϩ-8               	151488492	         7.954 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-Ϩ-8    	34277078	        35.30 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-Ϩ-8      	33704025	        35.82 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-Ϩ-8             	150425420	         7.969 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-✐-8         	49294844	        23.75 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-✐-8           	34984142	        34.59 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-✐-8             	133667400	         8.989 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-✐-8               	152579082	         7.817 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-✐-8    	34333825	        36.21 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-✐-8      	33334053	        35.20 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-✐-8             	152239486	         7.901 ns/op	       1 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8          	 4190718	       277.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8          	22362663	        52.89 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-100-8         	19529896	        61.01 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	11137023	       105.2 ns/op	     136 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-100-8        	33179738	        37.13 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	31806244	        37.18 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	  443863	      2672 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	 4410189	       272.4 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	 4323145	       274.1 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	 3881739	       312.0 ns/op	    1048 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 4364182	       271.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 4410223	       270.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	   45091	     26577 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	  653904	      1878 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  671106	      1804 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	  634357	      1923 ns/op	   10264 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  429177	      2728 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  441177	      2676 ns/op	       0 B/op	       0 allocs/op
BenchmarkMemoryOperations/Alloc-64-8        	20982091	        56.21 ns/op	      64 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-1024-8      	 4438129	       272.5 ns/op	    1024 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-4096-8      	 1200166	       989.7 ns/op	    4096 B/op	       1 allocs/op
BenchmarkConcurrentOperations-8             	 4919564	       265.8 ns/op	     120 B/op	       3 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	 8119792	       159.8 ns/op	     104 B/op	       3 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	   10000	   1765210 ns/op	 5120627 B/op	    5002 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	     121	  11526390 ns/op	63963465 B/op	      63 allocs/op
BenchmarkGet/ChunkPipe-Get-100-8            	100000000	        11.70 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-100-8                	165742989	         7.262 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-1000-8           	97296187	        11.89 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-1000-8               	157479946	         7.479 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-10000-8          	99026160	        11.93 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-10000-8              	162235918	         7.414 ns/op	       0 B/op	       0 allocs/op
```