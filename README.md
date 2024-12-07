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
BenchmarkPush10x10000-8            	    2079	    575637 ns/op	 1179896 B/op	       9 allocs/op
BenchmarkPush100x1000-8            	   24954	     48259 ns/op	  131273 B/op	       5 allocs/op
BenchmarkPush1000x100-8            	   61168	     19348 ns/op	  131272 B/op	       5 allocs/op
BenchmarkPush10000x10-8            	   71846	     16786 ns/op	  131273 B/op	       5 allocs/op
BenchmarkPopEnd10x10000-8          	     392	   3100202 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd100x1000-8          	     390	   3097951 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd1000x100-8          	     369	   3165706 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopEnd10000x10-8          	     388	   3123532 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkEnd10x10000-8     	    3909	    308481 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkEnd100x1000-8     	   32911	     36155 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkEnd1000x100-8     	  352492	      3430 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopFront10x10000-8        	     388	   3015314 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopFront100x1000-8        	     381	   3049988 ns/op	       2 B/op	       0 allocs/op
BenchmarkPopFront1000x100-8        	     376	   3064213 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopFront10000x10-8        	     387	   3106323 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkFront10x10000-8   	    3810	    312856 ns/op	       0 B/op	       0 allocs/op
BenchmarkPopChunkFront100x1000-8   	   32574	     36609 ns/op	       1 B/op	       0 allocs/op
BenchmarkPopChunkFront1000x100-8   	  341905	      3494 ns/op	       1 B/op	       0 allocs/op
BenchmarkGet10x10000-8             	     454	   2620929 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet100x1000-8             	     699	   1701275 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet1000x100-8             	     963	   1257824 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet10000x10-8             	    1461	    816894 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8 	 2023226	       584.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8 	251539732	         4.766 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-100-8         	 1281855	       926.4 ns/op	    4121 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	   94732	     12872 ns/op	   98369 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-100-8        	37124523	        32.60 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	36359253	        32.48 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	  208170	      5735 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	253013617	         4.799 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	  818180	      1377 ns/op	    4121 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	   92185	     12994 ns/op	   98369 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 4433034	       271.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 4455002	       269.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	   20959	     57489 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	251453919	         4.752 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  150056	      7970 ns/op	   10332 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	   95544	     12680 ns/op	   98369 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  451576	      2655 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  446856	      2663 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentOperations-8             	 3877998	       282.4 ns/op	     183 B/op	       1 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	12486856	        96.44 ns/op	      23 B/op	       1 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	  106574	    113315 ns/op	     163 B/op	       0 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	  110608	    118490 ns/op	     157 B/op	       0 allocs/op
PASS
coverage: 87.3% of statements
ok  	github.com/HazelnutParadise/go-chunkpipe	178.874s
```