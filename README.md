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
BenchmarkPush10x10000-8            	    1975	    588818 ns/op	 1471250 B/op	      22 allocs/op
BenchmarkPush100x1000-8            	   27477	     42939 ns/op	   70416 B/op	      14 allocs/op
BenchmarkPush1000x100-8            	  242695	      5063 ns/op	    9488 B/op	      11 allocs/op
BenchmarkPush10000x10-8            	 1541936	       784.1 ns/op	    1168 B/op	       8 allocs/op
BenchmarkPopEnd10x10000-8          	     393	   3063405 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopEnd100x1000-8          	     390	   3103725 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd1000x100-8          	     380	   3086392 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd10000x10-8          	     394	   3147257 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd10x10000-8     	    3658	    327186 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd100x1000-8     	   35132	     34392 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd1000x100-8     	  357068	      3396 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10x10000-8        	     375	   3182231 ns/op	       2 B/op	       0 allocs/op
BenchmarkPopFront100x1000-8        	     372	   3132868 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront1000x100-8        	     384	   3126844 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10000x10-8        	     370	   3124347 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront10x10000-8   	    3400	    349421 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkFront100x1000-8   	   33212	     36146 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront1000x100-8   	  335590	      3633 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10x10000-8             	     294	   3917424 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet100x1000-8             	     362	   3182840 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet1000x100-8             	     441	   2773150 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10000x10-8             	     532	   2345628 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8 	  343912	      3473 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8 	243476312	         4.895 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-100-8         	 7504990	       158.1 ns/op	     136 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	13200513	        91.09 ns/op	      48 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-100-8        	35556332	        33.60 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	35357124	        33.89 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	   31726	     37196 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	242948145	         4.892 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	 1474521	       798.5 ns/op	    1048 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	13232380	        91.34 ns/op	      48 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 4268137	       279.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 4317121	       280.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	    3254	    409494 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	245612413	         4.940 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  157148	      7648 ns/op	   10267 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	12230442	        90.33 ns/op	      48 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  436372	      2733 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  439077	      2734 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentOperations-8             	 4058548	       298.1 ns/op	     185 B/op	       1 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	12251431	        95.56 ns/op	      24 B/op	       1 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	  107380	    118717 ns/op	     165 B/op	       0 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	  109278	    121646 ns/op	     162 B/op	       0 allocs/op
PASS
coverage: 81.7% of statements
ok  	github.com/HazelnutParadise/go-chunkpipe	160.519s

```
