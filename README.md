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
BenchmarkPush/ChunkPipe-10-8      	10189221	       103.2 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-10-8          	100000000	        51.28 ns/op	      57 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100-8     	14097289	        81.52 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-100-8         	12567604	       421.4 ns/op	     570 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000-8    	14859344	        97.84 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-1000-8        	 1000000	      9355 ns/op	    5736 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-10000-8   	12055174	        97.14 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-10000-8       	  179115	     61552 ns/op	   50044 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100000-8  	 4847977	       232.7 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-100000-8      	   10000	    963146 ns/op	  551085 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000000-8 	 7635669	       159.9 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-1000000-8     	    1258	   5544641 ns/op	 5092598 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-10-8         	22404010	        44.76 ns/op	       6 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-10-8           	27219898	        86.51 ns/op	       6 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-10-8             	110151345	        10.92 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-10-8               	144679620	         9.592 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-10-8    	26428676	        46.78 ns/op	       8 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-10-8      	26556116	        45.23 ns/op	       8 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-10-8             	120948088	        10.03 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-100-8        	43880588	        26.40 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-100-8          	32386393	        37.65 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-100-8            	131540048	         9.132 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-100-8              	150495794	         8.112 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-100-8   	33588543	        35.60 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-100-8     	33217774	        36.85 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-100-8            	144687132	         8.471 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-1000-8       	47547288	        24.35 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-1000-8         	34042070	        36.63 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-1000-8           	131296255	         9.141 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-1000-8             	151756510	         7.898 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-1000-8  	34486960	        34.88 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-1000-8    	32362987	        35.14 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-1000-8           	151459046	         7.900 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-10000-8      	48294686	        24.19 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-10000-8        	32628091	        35.25 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-10000-8          	134207332	         8.918 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-10000-8            	154317309	         7.951 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-10000-8 	34487523	        34.92 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-10000-8   	34341626	        36.06 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-10000-8          	153412496	         7.982 ns/op	       1 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8           	 4315234	       277.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8           	22972132	        50.98 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-100-8          	20139468	        60.48 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-100-8          	11185042	       104.2 ns/op	     136 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-100-8         	32052748	        37.62 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8    	31169704	        37.82 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8          	  451890	      2709 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8          	 4880830	       248.7 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-1000-8         	 4726509	       251.5 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-1000-8         	 4230631	       291.0 ns/op	    1048 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-1000-8        	 4479474	       270.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8   	 4425966	       274.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8         	   45409	     26992 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8         	  638883	      1753 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-10000-8        	  700612	      1907 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-10000-8        	  645282	      1848 ns/op	   10264 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-10000-8       	  442518	      2709 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8  	  442296	      2659 ns/op	       0 B/op	       0 allocs/op
BenchmarkMemoryOperations/Alloc-64-8         	18688429	        63.43 ns/op	      64 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-1024-8       	 4779922	       258.5 ns/op	    1024 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-4096-8       	 1270318	       940.4 ns/op	    4096 B/op	       1 allocs/op
BenchmarkConcurrentOperations-8              	 4185398	       305.1 ns/op	     120 B/op	       3 allocs/op
BenchmarkMixedOperations/ChunkPipe-8         	 7245939	       176.2 ns/op	     104 B/op	       3 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8        	   10000	   1802789 ns/op	 5120647 B/op	    5002 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8     	     140	  11872449 ns/op	73927178 B/op	      85 allocs/op
BenchmarkGet/ChunkPipe-Get-100-8             	100000000	        11.65 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-100-8                 	163230788	         7.269 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-1000-8            	100000000	        11.46 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-1000-8                	165678051	         7.414 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-10000-8           	100000000	        11.49 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-10000-8               	160955558	         7.571 ns/op	       0 B/op	       0 allocs/op
BenchmarkPushBatch/ChunkPipe-Push-100-8      	  698728	      1703 ns/op	    3177 B/op	      16 allocs/op
BenchmarkPushBatch/Slice-Append-100-8        	12299889	        99.24 ns/op	     453 B/op	       0 allocs/op
BenchmarkPushBatch/ChunkPipe-Push-1000-8     	  693822	      1655 ns/op	    3177 B/op	      16 allocs/op
BenchmarkPushBatch/Slice-Append-1000-8       	 1420944	       811.1 ns/op	    5094 B/op	       0 allocs/op
BenchmarkPushBatch/ChunkPipe-Push-10000-8    	  694972	      1618 ns/op	    3177 B/op	      16 allocs/op
BenchmarkPushBatch/Slice-Append-10000-8      	  127490	      9292 ns/op	   51606 B/op	       0 allocs/op
BenchmarkPushBatch/ChunkPipe-Push-100000-8   	  709759	      1636 ns/op	    3177 B/op	      16 allocs/op
BenchmarkPushBatch/Slice-Append-100000-8     	   13843	     82793 ns/op	  494389 B/op	       0 allocs/op
PASS
coverage: 74.5% of statements
ok  	github.com/HazelnutParadise/go-chunkpipe	182.315s
```