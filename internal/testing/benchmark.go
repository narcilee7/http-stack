package testing

import (
	"fmt"
	"testing"
	"time"
)

// BenchSuite 基准测试套件
type BenchSuite struct {
	Name      string
	SetupFunc func(*testing.B) interface{}
	TestFunc  func(*testing.B, interface{})
	Cleanup   func(interface{})
}

// RunBenchSuite 运行基准测试套件
func RunBenchSuite(b *testing.B, suite BenchSuite) {
	// 标记当前函数为测试辅助函数
	b.Helper()

	var setupData interface{}

	if suite.SetupFunc != nil {
		// 如果SetupFunc不为nil, 则调用它来获取测试数据
		setupData = suite.SetupFunc(b)
	}

	defer func() {
		// 在测试函数执行完毕后, 调用Cleanup函数来清理测试数据
		if suite.Cleanup != nil && setupData != nil {
			suite.Cleanup(setupData)
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		suite.TestFunc(b, setupData)
	}
}

// BenchmarkConfig 基准测试配置
type BenchmarkConfig struct {
	MinDuration time.Duration
	MaxDuration time.Duration
	Memory      bool
	CPU         bool
}

// DefaultBenchmarkConfig 默认基准测试配置
func DefaultBenchmarkConfig() BenchmarkConfig {
	return BenchmarkConfig{
		MinDuration: 1 * time.Second,
		MaxDuration: 10 * time.Second,
		Memory:      true,
		CPU:         false,
	}
}

// BenchmarkRunner 基准测试执行器
type BenchmarkRunner struct {
	config BenchmarkConfig
}

// NewBenchmarkRunner 创建基准测试执行器
func NewBenchmarkRunner(config BenchmarkConfig) *BenchmarkRunner {
	return &BenchmarkRunner{config: config}
}

// Run 运行基准测试
func (br *BenchmarkRunner) Run(b *testing.B, name string, fn func(*testing.B)) {
	b.Helper()

	if br.config.Memory {
		b.ReportAllocs()
	}

	b.Run(name, fn)
}

// MemoryBenchmark 内存基准测试
func MemoryBenchmark(b *testing.B, name string, fn func(*testing.B)) {
	b.Helper()
	b.ReportAllocs()
	b.Run(name, fn)
}

// CPUBenchmark CPU基准测试
func CPUBenchmark(b *testing.B, name string, fn func(*testing.B)) {
	b.Helper()
	b.Run(name, fn)
}

// ConcurrentBenchmark 并发基准测试
func ConcurrentBenchmark(b *testing.B, name string, goroutines int, fn func(*testing.B)) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		b.SetParallelism(goroutines)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				fn(b)
			}
		})
	})
}

// ThroughputBenchmark 吞吐量基准测试
func ThroughputBenchmark(b *testing.B, name string, bytesPerOp int64, fn func(*testing.B)) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		b.SetBytes(bytesPerOp)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fn(b)
		}
	})
}

// LatencyBenchmark 延迟基准测试
func LatencyBenchmark(b *testing.B, name string, fn func() time.Duration) {
	b.Helper()
	b.Run(name, func(b *testing.B) {
		var totalDuration time.Duration
		for i := 0; i < b.N; i++ {
			totalDuration += fn()
		}

		avgLatency := totalDuration / time.Duration(b.N)
		b.ReportMetric(float64(avgLatency.Nanoseconds()), "ns/op")
	})
}

// ComparisonBenchmark 对比基准测试
func ComparisonBenchmark(b *testing.B, tests map[string]func(*testing.B)) {
	b.Helper()

	for name, fn := range tests {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				fn(b)
			}
		})
	}
}

// ProgressiveBenchmark 渐进式基准测试
func ProgressiveBenchmark(b *testing.B, name string, sizes []int, fn func(*testing.B, int)) {
	b.Helper()

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%s/size_%d", name, size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				fn(b, size)
			}
		})
	}
}

// WarmupBenchmark 预热基准测试
func WarmupBenchmark(b *testing.B, name string, warmupRuns int, fn func(*testing.B)) {
	b.Helper()

	b.Run(name, func(b *testing.B) {
		// 预热阶段
		for i := 0; i < warmupRuns; i++ {
			fn(b)
		}

		b.ResetTimer()
		b.ReportAllocs()

		// 正式测试
		for i := 0; i < b.N; i++ {
			fn(b)
		}
	})
}
