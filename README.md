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
BenchmarkPush10x10000-8            	    1770	    599814 ns/op	 1471136 B/op	      20 allocs/op
BenchmarkPush100x1000-8            	   27283	     43952 ns/op	   70304 B/op	      12 allocs/op
BenchmarkPush1000x100-8            	  241518	      5049 ns/op	    9376 B/op	       9 allocs/op
BenchmarkPush10000x10-8            	 1680913	       719.7 ns/op	    1056 B/op	       6 allocs/op
BenchmarkPopEnd10x10000-8          	     381	   3234456 ns/op	       4 B/op	       0 allocs/op
BenchmarkPopEnd100x1000-8          	     378	   3133658 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopEnd1000x100-8          	     378	   3160448 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd10000x10-8          	     382	   3185949 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd10x10000-8     	    3572	    327240 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkEnd100x1000-8     	   35750	     33905 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd1000x100-8     	  356383	      3366 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10x10000-8        	     368	   3159027 ns/op	       2 B/op	       0 allocs/op
BenchmarkPopFront100x1000-8        	     375	   3166341 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront1000x100-8        	     380	   3151978 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10000x10-8        	     381	   3138413 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront10x10000-8   	    3265	    349905 ns/op	       2 B/op	       0 allocs/op
BenchmarkPopChunkFront100x1000-8   	   31935	     37352 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront1000x100-8   	  332476	      3623 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10x10000-8             	     304	   4184874 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet100x1000-8             	     373	   3210656 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet1000x100-8             	     433	   2753725 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10000x10-8             	     526	   2253921 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8 	  342406	      3509 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8 	246001219	         4.901 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-100-8         	10428896	       114.6 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	24748742	        48.19 ns/op	      24 B/op	       1 allocs/op
BenchmarkIterators/NativeSlice-100-8        	35795956	        34.11 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	36046534	        33.56 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	   34440	     34977 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	246004669	         4.921 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	 1612730	       744.9 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	24778080	        48.39 ns/op	      24 B/op	       1 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 4338097	       278.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 4283628	       277.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	    3286	    347558 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	243000354	         4.897 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  161496	      7452 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	24733107	        49.11 ns/op	      24 B/op	       1 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  438436	      2724 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  437886	      2730 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentOperations-8             	 3940384	       313.7 ns/op	     202 B/op	       1 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	12197779	       120.5 ns/op	      24 B/op	       1 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	   75008	     87122 ns/op	     187 B/op	       0 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	  108927	    122704 ns/op	     163 B/op	       0 allocs/op
PASS
coverage: 86.0% of statements
ok  	github.com/HazelnutParadise/go-chunkpipe	154.098s
```