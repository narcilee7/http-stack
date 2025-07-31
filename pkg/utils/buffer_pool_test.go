package utils

import (
	"bytes"
	"testing"

	testhelper "github.com/narcilee7/http-stack/internal/testing"
)

func TestBufferPool(t *testing.T) {
	h := testhelper.NewHelper(t)

	t.Run("basic_operations", func(t *testing.T) {
		pool := NewBufferPool()
		h.AssertNotNil(pool, "BufferPool should not be nil")

		// 获取缓冲区
		buf := pool.Get()
		h.AssertNotNil(buf, "Buffer should not be nil")
		h.AssertEqual(buf.Len(), 0, "New buffer should be empty")

		// 写入数据
		testData := []byte("hello world")
		buf.Write(testData)
		h.AssertEqual(buf.Len(), len(testData), "Buffer length should match written data")

		// 读取数据
		readData := buf.Bytes()
		h.AssertEqual(string(readData), string(testData), "Read data should match written data")

		// 归还缓冲区
		pool.Put(buf)
		h.AssertEqual(buf.Len(), 0, "Buffer should be reset after returning to pool")
	})

	t.Run("pool_reuse", func(t *testing.T) {
		pool := NewBufferPool()

		// 获取第一个缓冲区
		buf1 := pool.Get()
		buf1.WriteString("test1")
		originalPtr := buf1

		// 归还缓冲区
		pool.Put(buf1)

		// 再次获取缓冲区，应该是同一个被重用的
		buf2 := pool.Get()
		h.AssertEqual(buf2, originalPtr, "Should reuse the same buffer")
		h.AssertEqual(buf2.Len(), 0, "Reused buffer should be clean")
	})

	t.Run("concurrent_access", func(t *testing.T) {
		pool := NewBufferPool()
		const numGoroutines = 100
		const numOperations = 1000

		done := make(chan bool, numGoroutines)

		// 并发获取和归还缓冲区
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < numOperations; j++ {
					buf := pool.Get()
					buf.WriteString("concurrent test")
					pool.Put(buf)
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// 验证pool状态正常
		buf := pool.Get()
		h.AssertNotNil(buf, "Pool should still work after concurrent access")
		h.AssertEqual(buf.Len(), 0, "Buffer should be clean")
		pool.Put(buf)
	})

	t.Run("size_based_pools", func(t *testing.T) {
		pool := NewBufferPoolWithSizes([]int{64, 256, 1024})

		// 小缓冲区
		smallBuf := pool.GetWithSize(32)
		h.AssertTrue(smallBuf.Cap() >= 32, "Buffer capacity should be at least requested size")

		// 中等缓冲区
		mediumBuf := pool.GetWithSize(128)
		h.AssertTrue(mediumBuf.Cap() >= 128, "Buffer capacity should be at least requested size")

		// 大缓冲区
		largeBuf := pool.GetWithSize(512)
		h.AssertTrue(largeBuf.Cap() >= 512, "Buffer capacity should be at least requested size")

		pool.Put(smallBuf)
		pool.Put(mediumBuf)
		pool.Put(largeBuf)
	})

	t.Run("auto_grow", func(t *testing.T) {
		pool := NewBufferPool()
		buf := pool.Get()

		// 写入大量数据测试自动扩容
		largeData := make([]byte, 10*1024) // 10KB
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		buf.Write(largeData)
		h.AssertEqual(buf.Len(), len(largeData), "Buffer should contain all written data")
		h.AssertTrue(buf.Cap() >= len(largeData), "Buffer capacity should grow")

		pool.Put(buf)
	})

	t.Run("memory_efficiency", func(t *testing.T) {
		pool := NewBufferPool()

		// 测试内存重用效率
		buffers := make([]*bytes.Buffer, 10)
		for i := 0; i < 10; i++ {
			buffers[i] = pool.Get()
			buffers[i].WriteString("memory test")
		}

		// 归还所有缓冲区
		for _, buf := range buffers {
			pool.Put(buf)
		}

		// 再次获取，应该重用之前的缓冲区
		newBuf := pool.Get()
		h.AssertNotNil(newBuf, "Should get buffer from pool")
		h.AssertEqual(newBuf.Len(), 0, "Reused buffer should be clean")

		pool.Put(newBuf)
	})
}

func TestBufferPoolStats(t *testing.T) {
	h := testhelper.NewHelper(t)

	t.Run("pool_statistics", func(t *testing.T) {
		pool := NewBufferPool()

		// 初始统计
		stats := pool.Stats()
		h.AssertEqual(stats.TotalGets, int64(0), "Initial gets should be 0")
		h.AssertEqual(stats.TotalPuts, int64(0), "Initial puts should be 0")

		// 执行一些操作
		buf1 := pool.Get()
		buf2 := pool.Get()
		pool.Put(buf1)

		stats = pool.Stats()
		h.AssertEqual(stats.TotalGets, int64(2), "Gets should be 2")
		h.AssertEqual(stats.TotalPuts, int64(1), "Puts should be 1")

		pool.Put(buf2)
		stats = pool.Stats()
		h.AssertEqual(stats.TotalPuts, int64(2), "Puts should be 2")
	})
}

// 基准测试
func BenchmarkBufferPool(b *testing.B) {
	pool := NewBufferPool()

	testhelper.MemoryBenchmark(b, "buffer_get_put", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := pool.Get()
			buf.WriteString("benchmark test data")
			pool.Put(buf)
		}
	})
}

func BenchmarkBufferPoolConcurrent(b *testing.B) {
	pool := NewBufferPool()

	testhelper.ConcurrentBenchmark(b, "concurrent_buffer_ops", 10, func(b *testing.B) {
		buf := pool.Get()
		buf.WriteString("concurrent benchmark")
		pool.Put(buf)
	})
}

func BenchmarkBufferPoolVsDirectAllocation(b *testing.B) {
	pool := NewBufferPool()

	testhelper.ComparisonBenchmark(b, map[string]func(*testing.B){
		"pool": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf := pool.Get()
				buf.WriteString("pool allocation test")
				pool.Put(buf)
			}
		},
		"direct": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf := &bytes.Buffer{}
				buf.WriteString("direct allocation test")
			}
		},
	})
}

func BenchmarkBufferPoolSizes(b *testing.B) {
	pool := NewBufferPoolWithSizes([]int{64, 256, 1024, 4096})
	sizes := []int{32, 128, 512, 2048}

	testhelper.ProgressiveBenchmark(b, "buffer_sizes", sizes, func(b *testing.B, size int) {
		for i := 0; i < b.N; i++ {
			buf := pool.GetWithSize(size)
			data := make([]byte, size)
			buf.Write(data)
			pool.Put(buf)
		}
	})
}
