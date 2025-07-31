package testing

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// Helper 提供常用的测试辅助功能
type Helper struct {
	t *testing.T
}

// NewHelper 创建新的测试辅助器
func NewHelper(t *testing.T) *Helper {
	return &Helper{t: t}
}

// AssertEqual 断言两个值相等
func (h *Helper) AssertEqual(got, want interface{}, msgAndArgs ...interface{}) {
	h.t.Helper()
	// 利用反射判断两个值是否相等
	if !reflect.DeepEqual(got, want) {
		msg := fmt.Sprintf("AssertEqual failed:\ngot:  %v\nwant: %v", got, want)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
			msg = fmt.Sprintf("%s\n%s", msg, msg)
		}
		h.t.Fatal(msg)
	}
}

// AssertNotEqual 断言两个值不相等
func (h *Helper) AssertNotEqual(got, notWant interface{}, msgAndArgs ...interface{}) {
	h.t.Helper()
	if reflect.DeepEqual(got, notWant) {
		msg := fmt.Sprintf("AssertNotEqual failed:\ngot:      %v\nnotWant:  %v", got, notWant)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
			msg = fmt.Sprintf("%s\n%s", msg, msg)
		}
		h.t.Fatal(msg)
	}
}

// AssertNil 断言值为nil
func (h *Helper) AssertNil(got interface{}, msgAndArgs ...interface{}) {
	h.t.Helper()
	if !isNil(got) {
		msg := fmt.Sprintf("AssertNil failed:\ngot: %v", got)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
			msg = fmt.Sprintf("%s\n%s", msg, msg)
		}
		h.t.Fatal(msg)
	}
}

// AssertNotNil 断言值不为nil
func (h *Helper) AssertNotNil(got interface{}, msgAndArgs ...interface{}) {
	h.t.Helper()
	if isNil(got) {
		msg := "AssertNotNil failed: got nil"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		h.t.Fatal(msg)
	}
}

// AssertTrue 断言值为true
func (h *Helper) AssertTrue(got bool, msgAndArgs ...interface{}) {
	h.t.Helper()
	if !got {
		msg := "AssertTrue failed: got false"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		h.t.Fatal(msg)
	}
}

// AssertFalse 断言值为false
func (h *Helper) AssertFalse(got bool, msgAndArgs ...interface{}) {
	h.t.Helper()
	if got {
		msg := "AssertFalse failed: got true"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		h.t.Fatal(msg)
	}
}

// AssertPanic 断言函数会发生panic
func (h *Helper) AssertPanic(fn func(), msgAndArgs ...interface{}) {
	h.t.Helper()
	defer func() {
		if r := recover(); r == nil {
			msg := "AssertPanic failed: function did not panic"
			if len(msgAndArgs) > 0 {
				msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
			}
			h.t.Fatal(msg)
		}
	}()
	fn()
}

// AssertNoPanic 断言函数不会发生panic
func (h *Helper) AssertNoPanic(fn func(), msgAndArgs ...interface{}) {
	h.t.Helper()
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("AssertNoPanic failed: function panicked with: %v", r)
			if len(msgAndArgs) > 0 {
				msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
				msg = fmt.Sprintf("%s\n%s", msg, msg)
			}
			h.t.Fatal(msg)
		}
	}()
	fn()
}

// AssertDuration 断言时间耗时在指定范围内
func (h *Helper) AssertDuration(fn func(), min, max time.Duration, msgAndArgs ...interface{}) {
	h.t.Helper()
	start := time.Now()
	fn()
	elapsed := time.Since(start)

	if elapsed < min || elapsed > max {
		msg := fmt.Sprintf("AssertDuration failed:\nelapsed: %v\nmin:     %v\nmax:     %v", elapsed, min, max)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
			msg = fmt.Sprintf("%s\n%s", msg, msg)
		}
		h.t.Fatal(msg)
	}
}

// isNil 检查值是否为nil
func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Interface:
		// 对于这些类型, 需要检查是否为nil
		return rv.IsNil()
	}
	// 对于其他类型, 直接返回false
	return false
}

// BenchmarkHelper 基准测试辅助器
type BenchmarkHelper struct {
	b *testing.B
}

// NewBenchmarkHelper 创建基准测试辅助器
func NewBenchmarkHelper(b *testing.B) *BenchmarkHelper {
	return &BenchmarkHelper{b: b}
}

// ResetTimer 重置计时器
func (bh *BenchmarkHelper) ResetTimer() {
	bh.b.ResetTimer()
}

// StopTimer 停止计时器
func (bh *BenchmarkHelper) StopTimer() {
	bh.b.StopTimer()
}

// StartTimer 开始计时器
func (bh *BenchmarkHelper) StartTimer() {
	bh.b.StartTimer()
}

// SetBytes 设置每次操作的字节数
func (bh *BenchmarkHelper) SetBytes(n int64) {
	bh.b.SetBytes(n)
}

// ReportAllocs 报告内存分配情况
func (bh *BenchmarkHelper) ReportAllocs() {
	bh.b.ReportAllocs()
}
