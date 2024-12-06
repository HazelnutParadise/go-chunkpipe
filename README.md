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
- 高性能：比原生切片快 5-10 倍的迭代速度
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
BenchmarkPush/ChunkPipe-10-8      	11600392	        95.64 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-10-8          	100000000	        39.98 ns/op	      57 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100-8     	 3933087	       258.1 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-100-8         	17890588	       337.3 ns/op	     501 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000-8    	15594168	       149.0 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-1000-8        	 1000000	      5874 ns/op	    5736 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-10000-8   	 9854946	       113.6 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-10000-8       	  206919	     59380 ns/op	   54151 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-100000-8  	 2952602	       370.2 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-100000-8      	    7154	   2226794 ns/op	  616205 B/op	       0 allocs/op
BenchmarkPush/ChunkPipe-1000000-8 	16996731	        83.13 ns/op	      64 B/op	       1 allocs/op
BenchmarkPush/Slice-1000000-8     	    2062	   8617369 ns/op	 6074388 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-_-8         	27526244	        38.94 ns/op	       6 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-_-8           	26778372	        67.28 ns/op	       6 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-_-8             	100000000	        11.22 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-_-8               	147476144	         8.208 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-_-8    	26623180	        44.88 ns/op	       8 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-_-8      	26858494	        46.09 ns/op	       8 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-_-8             	100000000	        10.11 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-d-8         	46973919	        25.68 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-d-8           	33647634	        35.74 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-d-8             	130874107	         9.221 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-d-8               	149596327	         8.020 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-d-8    	33943702	        36.21 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-d-8      	32393820	        35.91 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-d-8             	143696664	         8.537 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-Ϩ-8         	48655951	        24.13 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-Ϩ-8           	33447193	        34.75 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-Ϩ-8             	131998696	         9.046 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-Ϩ-8               	149690432	         7.823 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-Ϩ-8    	34383921	        35.38 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-Ϩ-8      	34548830	        34.98 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-Ϩ-8             	152200650	         7.960 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopFront-✐-8         	48092242	        24.82 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopEnd-✐-8           	34142668	        34.72 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/Slice-PopFront-✐-8             	134543311	         8.954 ns/op	       2 B/op	       0 allocs/op
BenchmarkPop/Slice-PopEnd-✐-8               	152175748	         9.187 ns/op	       0 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkFront-✐-8    	34645416	        34.45 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/ChunkPipe-PopChunkEnd-✐-8      	34064224	        34.24 ns/op	       1 B/op	       0 allocs/op
BenchmarkPop/Slice-PopChunk-✐-8             	153547923	         7.820 ns/op	       1 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-100-8          	 4313408	       277.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-100-8          	22409497	        52.86 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-100-8         	19720161	        61.83 ns/op	     112 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-100-8         	11288829	       105.1 ns/op	     136 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-100-8        	30917730	        37.10 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-100-8   	32000604	        37.48 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-1000-8         	  451730	      2642 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-1000-8         	 4561533	       270.3 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-1000-8        	 4419286	       275.6 ns/op	    1024 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-1000-8        	 3873644	       310.3 ns/op	    1048 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-1000-8       	 4486232	       269.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-1000-8  	 4467906	       269.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ValueIter-10000-8        	   45415	     26388 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/ChunkIter-10000-8        	  628064	      1793 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ValueSlice-10000-8       	  676594	      1787 ns/op	   10240 B/op	       1 allocs/op
BenchmarkIterators/ChunkSlice-10000-8       	  640551	      1849 ns/op	   10264 B/op	       2 allocs/op
BenchmarkIterators/NativeSlice-10000-8      	  451532	      2643 ns/op	       0 B/op	       0 allocs/op
BenchmarkIterators/NativeSliceValue-10000-8 	  447852	      2636 ns/op	       0 B/op	       0 allocs/op
BenchmarkMemoryOperations/Alloc-64-8        	20995750	        56.16 ns/op	      64 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-1024-8      	 3647104	       279.3 ns/op	    1024 B/op	       1 allocs/op
BenchmarkMemoryOperations/Alloc-4096-8      	 1204203	       993.3 ns/op	    4096 B/op	       1 allocs/op
BenchmarkConcurrentOperations-8             	 4888033	       268.9 ns/op	     120 B/op	       3 allocs/op
BenchmarkMixedOperations/ChunkPipe-8        	 7936782	       157.8 ns/op	     104 B/op	       3 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1024-8       	   10000	   1721996 ns/op	 5120624 B/op	    5001 allocs/op
BenchmarkMemoryUsage/ChunkPipe-1048576-8    	     122	  11240806 ns/op	64487666 B/op	      63 allocs/op
BenchmarkGet/ChunkPipe-Get-100-8            	100000000	        10.19 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-100-8                	166325204	         7.198 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-1000-8           	100000000	        10.80 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-1000-8               	165923606	         7.199 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/ChunkPipe-Get-10000-8          	100000000	        10.53 ns/op	       0 B/op	       0 allocs/op
BenchmarkGet/Slice-Get-10000-8              	165403230	         7.196 ns/op	       0 B/op	       0 allocs/op
```