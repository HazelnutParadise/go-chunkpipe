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
BenchmarkPush10x10000-8            	    2013	    566362 ns/op	 1179943 B/op	       9 allocs/op
BenchmarkPush100x1000-8            	   25020	     47686 ns/op	  131304 B/op	       5 allocs/op
BenchmarkPush1000x100-8            	   62065	     19494 ns/op	  131304 B/op	       5 allocs/op
BenchmarkPush10000x10-8            	   71572	     23016 ns/op	  131308 B/op	       5 allocs/op
BenchmarkPopEnd10x10000-8          	      43	  27161126 ns/op	 2400015 B/op	  100000 allocs/op
BenchmarkPopEnd100x1000-8          	      45	  27165909 ns/op	 2400010 B/op	  100000 allocs/op
BenchmarkPopEnd1000x100-8          	      43	  27027142 ns/op	 2400006 B/op	  100000 allocs/op
BenchmarkPopEnd10000x10-8          	      43	  26942665 ns/op	 2400022 B/op	  100000 allocs/op
BenchmarkPopChunkEnd10x10000-8     	     436	   2754643 ns/op	  240000 B/op	   10000 allocs/op
BenchmarkPopChunkEnd100x1000-8     	    4276	    280787 ns/op	   24000 B/op	    1000 allocs/op
BenchmarkPopChunkEnd1000x100-8     	   41013	     29166 ns/op	    2400 B/op	     100 allocs/op
BenchmarkPopFront10x10000-8        	      43	  27059611 ns/op	 2400004 B/op	  100000 allocs/op
BenchmarkPopFront100x1000-8        	      44	  26805903 ns/op	 2400006 B/op	  100000 allocs/op
BenchmarkPopFront1000x100-8        	      43	  26874927 ns/op	 2400013 B/op	  100000 allocs/op
BenchmarkPopFront10000x10-8        	      44	  26826712 ns/op	 2400010 B/op	  100000 allocs/op
BenchmarkPopChunkFront10x10000-8   	     429	   2774738 ns/op	  240000 B/op	   10000 allocs/op
BenchmarkPopChunkFront100x1000-8   	    4214	    281339 ns/op	   24000 B/op	    1000 allocs/op
BenchmarkPopChunkFront1000x100-8   	   41475	     29254 ns/op	    2401 B/op	     100 allocs/op
BenchmarkGet10x10000-8             	     108	  10209163 ns/op	  265681 B/op	  103736 allocs/op
BenchmarkGet100x1000-8             	     105	  10704772 ns/op	  310757 B/op	  103848 allocs/op
BenchmarkGet1000x100-8             	     104	  10104994 ns/op	  294479 B/op	  102949 allocs/op
BenchmarkGet10000x10-8             	     120	   9421340 ns/op	  247194 B/op	   93078 allocs/op
BenchmarkIterators/ValueIter-100-8 	  395061	      2987 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8 	251557412	         4.799 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-100-8         	  808396	      1271 ns/op	    4120 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	   84187	     14408 ns/op	   98337 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-100-8        	36216015	        32.96 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	36076946	        32.82 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	   40189	     29738 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	250386902	         4.795 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	  595996	      1887 ns/op	    4120 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	   80797	     14691 ns/op	   98337 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 4415086	       271.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 4402735	       270.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	    3988	    297724 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	250923721	         4.771 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  130034	      9042 ns/op	   10344 B/op	       2 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	   81898	     14316 ns/op	   98337 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  444825	      2683 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  445143	      2665 ns/op	       0 B/op	       0 allocs/op
BenchmarkConcurrentOperations-8             	 2097249	       520.0 ns/op	     207 B/op	       2 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	 3204522	       361.7 ns/op	      47 B/op	       2 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	  111050	    118954 ns/op	     157 B/op	       0 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	  111682	    119512 ns/op	     156 B/op	       0 allocs/op
PASS
coverage: 85.2% of statements
ok  	github.com/HazelnutParadise/go-chunkpipe	103.484s
```