# ChunkPipe 塊聯管

塊聯管是一個高性能的 Go 管道資料結構，它將插入的每個切片視為一個塊，並將多個塊連結成一個管道。

塊聯管可以一次存入或取出單個元素或整個塊，非常適合作為整塊資料的儲存與操作（例如需要資料連續性的場景），也適合作為 Queue 或 Stack 使用。

塊聯管甚至可以儲存幾乎任何類型，包括 map 或 struct 等複雜類型。

> [!NOTE]
> 使用 map 或 struct 時，會將其視為一個整體，不會對其內部資料進行迭代。

## 系統要求

- Go 1.22.7 或更高版本
- 支援 x86-64 架構
- 支援 Linux/Windows/macOS

## 特點

- 泛型支援：可處理任意類型資料
- 天生併發安全：優化的鎖機制
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

1. 迭代元素
```go
iter := cp.ValueIter()
for iter.Next() {
    value := iter.V()
    // do something with value
}
```

2. 迭代塊
```go
iter := cp.ChunkIter()
for iter.Next() {
    chunk := iter.V()
    // do something with chunk
}
```

## 性能

```bash
goos: darwin
goarch: amd64
pkg: github.com/HazelnutParadise/go-chunkpipe
cpu: Intel(R) Core(TM) i5-8257U CPU @ 1.40GHz
BenchmarkPush10x10000-8            	    2134	    578435 ns/op	 1212727 B/op	      10 allocs/op
BenchmarkPush100x1000-8            	   23961	     50681 ns/op	  164088 B/op	       6 allocs/op
BenchmarkPush1000x100-8            	   52257	     23299 ns/op	  164088 B/op	       6 allocs/op
BenchmarkPush10000x10-8            	   56702	     21088 ns/op	  164088 B/op	       6 allocs/op
BenchmarkPopEnd10x10000-8          	     393	   3122828 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd100x1000-8          	     379	   3136863 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd1000x100-8          	     378	   3168414 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd10000x10-8          	     379	   3103229 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd10x10000-8     	    3867	    321536 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd100x1000-8     	   35509	     34693 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd1000x100-8     	  335427	      3549 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10x10000-8        	     391	   3065170 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront100x1000-8        	     387	   3059391 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront1000x100-8        	     392	   3101665 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10000x10-8        	     378	   3064869 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront10x10000-8   	    4183	    285115 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront100x1000-8   	   39030	     31932 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront1000x100-8   	  372652	      3273 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10x10000-8             	     321	   3713754 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet100x1000-8             	     393	   3061433 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet1000x100-8             	     465	   2590734 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10000x10-8             	     585	   2041628 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8 	  793610	      1515 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8 	238633328	         4.944 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-100-8         	 1218969	       987.1 ns/op	    4121 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	   85279	     13900 ns/op	   98369 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-100-8        	36018661	        33.66 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	35790132	        33.49 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	   81828	     14623 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	248385224	         4.839 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	  787173	      1772 ns/op	    4121 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	   90889	     13527 ns/op	   98369 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 4392418	       273.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 4380146	       274.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	    8082	    144507 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	250984966	         4.773 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  148506	      9833 ns/op	   10399 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	   70692	     14508 ns/op	   98370 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  449798	      2663 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  450709	      2671 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentOperations-8             	 4199420	       269.8 ns/op	     196 B/op	       1 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	11835981	        91.42 ns/op	      23 B/op	       1 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	  110796	    118106 ns/op	     158 B/op	       0 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	  110793	    118370 ns/op	     158 B/op	       0 allocs/op
PASS
coverage: 76.2% of statements
ok  	github.com/HazelnutParadise/go-chunkpipe	190.343s
```