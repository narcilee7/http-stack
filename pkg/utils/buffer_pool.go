package utils

import (
	"bytes"
	"sort"
	"sync"
	"sync/atomic"
)

// PoolStats 缓冲区池统计信息
type PoolStats struct {
	TotalGets int64 // 总获取次数
	TotalPuts int64 // 总归还次数
	PoolHits  int64 // 池命中次数
	PoolSize  int   // 当前池大小
}

// BufferPool 缓冲区池，提供高效的bytes.Buffer复用
type BufferPool struct {
	pools map[int]*sync.Pool // 按大小分组的池
	sizes []int              // 支持的大小
	stats PoolStats          // 统计信息
	mutex sync.RWMutex       // 保护统计信息
}

// NewBufferPool 创建新的缓冲区池，使用默认大小
func NewBufferPool() *BufferPool {
	defaultSizes := []int{64, 256, 1024, 4096, 16384}
	return NewBufferPoolWithSizes(defaultSizes)
}

// NewBufferPoolWithSizes 创建支持指定大小的缓冲区池
func NewBufferPoolWithSizes(sizes []int) *BufferPool {
	// 排序大小数组以便查找
	sortedSizes := make([]int, len(sizes))
	// 复制并排序
	copy(sortedSizes, sizes)
	sort.Ints(sortedSizes)
	// 创建池
	pools := make(map[int]*sync.Pool)
	for _, size := range sortedSizes {
		pools[size] = &sync.Pool{
			New: func() interface{} {
				buffer := make([]byte, 0, size)
				return bytes.NewBuffer(buffer)
			},
		}
	}

	return &BufferPool{
		pools: pools,
		sizes: sortedSizes,
	}
}

// Get 获取一个缓冲区
func (bp *BufferPool) Get() *bytes.Buffer {
	return bp.GetWithSize(0)
}

// GetWithSize 获取指定最小容量的缓冲区
func (bp *BufferPool) GetWithSize(minSize int) *bytes.Buffer {
	// 利用原子操作增加获取次数，保证统计信息的一致性
	atomic.AddInt64(&bp.stats.TotalGets, 1)

	// 找到合适的池大小
	poolSize := bp.findPoolSize(minSize)

	if pool, exists := bp.pools[poolSize]; exists {
		if buf := pool.Get().(*bytes.Buffer); buf != nil {
			atomic.AddInt64(&bp.stats.PoolHits, 1)
			buf.Reset() // 确保缓冲区是干净的
			return buf
		}
	}

	// 如果没有找到合适的池，创建新的缓冲区
	capacity := poolSize
	if capacity == 0 {
		capacity = 256 // 默认容量
	}
	buf := make([]byte, 0, capacity)
	return bytes.NewBuffer(buf)
}

// Put 归还缓冲区到池中
func (bp *BufferPool) Put(buf *bytes.Buffer) {
	if buf == nil {
		return
	}

	atomic.AddInt64(&bp.stats.TotalPuts, 1)

	// 清空缓冲区
	buf.Reset()

	// 找到合适的池
	capacity := buf.Cap()
	poolSize := bp.findPoolSize(capacity)

	if pool, exists := bp.pools[poolSize]; exists {
		// 如果缓冲区太大，不放回池中以避免内存浪费
		if capacity <= poolSize*2 {
			pool.Put(buf)
		}
	}
}

// findPoolSize 找到合适的池大小
func (bp *BufferPool) findPoolSize(minSize int) int {
	// 如果minSize小于等于0，并且有大小数组，则返回最小的池大小
	if minSize <= 0 && len(bp.sizes) > 0 {
		return bp.sizes[0]
	}

	for _, size := range bp.sizes {
		if size >= minSize {
			return size
		}
	}

	if len(bp.sizes) > 0 {
		return bp.sizes[len(bp.sizes)-1]
	}

	return minSize
}

// Stats 获取池的统计信息
func (bp *BufferPool) Stats() PoolStats {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	stats := PoolStats{
		TotalGets: atomic.LoadInt64(&bp.stats.TotalGets),
		TotalPuts: atomic.LoadInt64(&bp.stats.TotalPuts),
		PoolHits:  atomic.LoadInt64(&bp.stats.PoolHits),
	}

	// 计算当前池大小（估算）
	totalSize := 0
	for size, pool := range bp.pools {
		// 这是一个估算，因为sync.Pool没有提供精确的大小信息
		_ = pool
		_ = size
		// 无法准确获取sync.Pool的大小，所以暂时设为0
	}
	stats.PoolSize = totalSize

	return stats
}

// Reset 重置池的统计信息
func (bp *BufferPool) Reset() {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	atomic.StoreInt64(&bp.stats.TotalGets, 0)
	atomic.StoreInt64(&bp.stats.TotalPuts, 0)
	atomic.StoreInt64(&bp.stats.PoolHits, 0)
}

// 全局缓冲区池实例
var globalBufferPool = NewBufferPool()

// GetBuffer 从全局池获取缓冲区
func GetBuffer() *bytes.Buffer {
	return globalBufferPool.Get()
}

// GetBufferWithSize 从全局池获取指定大小的缓冲区
func GetBufferWithSize(size int) *bytes.Buffer {
	return globalBufferPool.GetWithSize(size)
}

// PutBuffer 将缓冲区归还到全局池
func PutBuffer(buf *bytes.Buffer) {
	globalBufferPool.Put(buf)
}

// BufferPoolStats 获取全局池的统计信息
func BufferPoolStats() PoolStats {
	return globalBufferPool.Stats()
}
