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

- 零分配：大部分操作實現 0 allocs/op
- 泛型支援：可處理任意類型資料
- 天生併發安全：優化的原子操作和鎖機制
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
BenchmarkPush10x10000-8            	    1884	    612501 ns/op	 1471136 B/op	      20 allocs/op
BenchmarkPush100x1000-8            	   27525	     44808 ns/op	   70304 B/op	      12 allocs/op
BenchmarkPush1000x100-8            	  237518	      5345 ns/op	    9376 B/op	       9 allocs/op
BenchmarkPush10000x10-8            	 1651042	       891.0 ns/op	    1056 B/op	       6 allocs/op
BenchmarkPopEnd10x10000-8          	     376	   3168971 ns/op	       2 B/op	       0 allocs/op
BenchmarkPopEnd100x1000-8          	     379	   3594939 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopEnd1000x100-8          	     385	   3275996 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd10000x10-8          	     386	   3164231 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkEnd10x10000-8     	    3558	    334320 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkEnd100x1000-8     	   34908	     35866 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd1000x100-8     	  349916	      3374 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10x10000-8        	     373	   3177475 ns/op	       2 B/op	       0 allocs/op
BenchmarkPopFront100x1000-8        	     381	   3123788 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront1000x100-8        	     378	   3152555 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10000x10-8        	     382	   3206864 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront10x10000-8   	    3422	    349055 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkFront100x1000-8   	   32845	     36651 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront1000x100-8   	  334903	      3650 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10x10000-8             	     298	   3942902 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet100x1000-8             	     373	   3194297 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet1000x100-8             	     433	   2746873 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10000x10-8             	     534	   2272732 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8 	  344040	      3757 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8 	246398516	         4.916 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-100-8         	10515300	       110.7 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	24687423	        48.08 ns/op	      24 B/op	       1 allocs/op
BenchmarkIterators/NativeSlice-100-8        	35658332	        33.50 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	35670078	        33.44 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	   34335	     34956 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	244125649	         4.928 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	 1615987	       750.8 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	24397630	        49.35 ns/op	      24 B/op	       1 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 4280170	       278.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 4287931	       278.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	    3421	    357919 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	245870836	         4.922 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  158400	      7570 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	24456826	        48.14 ns/op	      24 B/op	       1 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  438918	      2731 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  436215	      2746 ns/op	       0 B/op	       0 allocs/op
BenchmarkMemoryOperations/Alloc-64-8        	19352060	        61.81 ns/op	      64 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-1024-8      	 5546092	       216.9 ns/op	    1024 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-4096-8      	 1470105	       789.0 ns/op	    4096 B/op	       1 allocs/op
BenchmarkConcurrentOperations-8             	 4198396	       318.3 ns/op	     205 B/op	       1 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	12262790	        96.23 ns/op	      24 B/op	       1 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	  107269	    119188 ns/op	     165 B/op	       0 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	  107803	    120298 ns/op	     164 B/op	       0 allocs/op
PASS
coverage: 80.2% of statements
ok  	github.com/HazelnutParadise/go-chunkpipe	166.537s
```
