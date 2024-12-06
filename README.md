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
BenchmarkPush10x10000-8            	    1974	    586864 ns/op	 1471248 B/op	      22 allocs/op
BenchmarkPush100x1000-8            	   27574	     42988 ns/op	   70416 B/op	      14 allocs/op
BenchmarkPush1000x100-8            	  233354	      5047 ns/op	    9488 B/op	      11 allocs/op
BenchmarkPush10000x10-8            	 1528660	       780.8 ns/op	    1168 B/op	       8 allocs/op
BenchmarkPopEnd10x10000-8          	     392	   3051630 ns/op	       4 B/op	       0 allocs/op
BenchmarkPopEnd100x1000-8          	     394	   3043298 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd1000x100-8          	     388	   3107123 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd10000x10-8          	     391	   3069640 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd10x10000-8     	    3645	    325008 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkEnd100x1000-8     	   35226	     34207 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd1000x100-8     	  360807	      3357 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10x10000-8        	     375	   3208795 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopFront100x1000-8        	     390	   3104614 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront1000x100-8        	     382	   3111622 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront10000x10-8        	     385	   3118345 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront10x10000-8   	    3459	    352298 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkFront100x1000-8   	   33417	     36018 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront1000x100-8   	  332174	      3626 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10x10000-8             	     309	   3903466 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet100x1000-8             	     376	   3282774 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet1000x100-8             	     440	   2781244 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10000x10-8             	     535	   2526536 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8 	  342345	      3469 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8 	239480233	         4.921 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-100-8         	 8930395	       131.5 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	18898441	        64.16 ns/op	      24 B/op	       1 allocs/op
BenchmarkIterators/NativeSlice-100-8        	36066339	        34.06 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	35114096	        33.73 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	   31244	     37233 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	242143387	         4.899 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	 1450884	       821.0 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	18989221	        63.90 ns/op	      24 B/op	       1 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 4262838	       279.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 4270159	       279.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	    3230	    374869 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	245922144	         4.907 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  160245	      7507 ns/op	   10243 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	18917146	        62.46 ns/op	      24 B/op	       1 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  432609	      2744 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  443089	      2748 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentOperations-8             	 3748714	       293.2 ns/op	     198 B/op	       1 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	12126363	        95.74 ns/op	      24 B/op	       1 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	  107095	    119486 ns/op	     165 B/op	       0 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	  107215	    121289 ns/op	     165 B/op	       0 allocs/op
PASS
coverage: 82.2% of statements
ok  	github.com/HazelnutParadise/go-chunkpipe	159.874s
```