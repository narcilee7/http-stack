package testing

import (
	"testing"
	"time"
)

func TestHelper(t *testing.T) {
	h := NewHelper(t)

	t.Run("AssertEqual", func(t *testing.T) {
		h.AssertEqual(1, 1)
		h.AssertEqual("hello", "hello")
		h.AssertEqual([]int{1, 2, 3}, []int{1, 2, 3})
	})

	t.Run("AssertNotEqual", func(t *testing.T) {
		h.AssertNotEqual(1, 2)
		h.AssertNotEqual("hello", "world")
	})

	t.Run("AssertNil", func(t *testing.T) {
		var ptr *int
		h.AssertNil(ptr)
		h.AssertNil(nil)
	})

	t.Run("AssertNotNil", func(t *testing.T) {
		value := 42
		h.AssertNotNil(&value)
		h.AssertNotNil("hello")
	})

	t.Run("AssertTrue", func(t *testing.T) {
		h.AssertTrue(true)
		h.AssertTrue(1 == 1)
	})

	t.Run("AssertFalse", func(t *testing.T) {
		h.AssertFalse(false)
		h.AssertFalse(1 == 2)
	})
}

func TestHelperPanic(t *testing.T) {
	h := NewHelper(t)

	t.Run("AssertPanic", func(t *testing.T) {
		h.AssertPanic(func() {
			panic("test panic")
		})
	})

	t.Run("AssertNoPanic", func(t *testing.T) {
		h.AssertNoPanic(func() {
			// 正常执行，不panic
		})
	})
}

func TestHelperDuration(t *testing.T) {
	h := NewHelper(t)

	t.Run("AssertDuration", func(t *testing.T) {
		h.AssertDuration(func() {
			time.Sleep(10 * time.Millisecond)
		}, 5*time.Millisecond, 50*time.Millisecond)
	})
}

func TestBenchmarkHelper(t *testing.T) {
	// 这里只是验证接口，不运行实际的基准测试
	b := &testing.B{}
	bh := NewBenchmarkHelper(b)

	if bh == nil {
		t.Fatal("BenchmarkHelper should not be nil")
	}
}

func BenchmarkExample(b *testing.B) {
	bh := NewBenchmarkHelper(b)
	bh.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// 示例基准测试
		_ = make([]byte, 1024)
	}
}
