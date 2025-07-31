package utils

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// HTTP时间格式常量
	HTTPTimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
	RFC850Format   = "Monday, 02-Jan-06 15:04:05 GMT"
	ASCTimeFormat  = "Mon Jan _2 15:04:05 2006"
)

// FormatHTTPTime 将时间格式化为HTTP标准格式 (RFC 7231)
func FormatHTTPTime(t time.Time) string {
	return t.UTC().Format(HTTPTimeFormat)
}

// ParseHTTPTime 解析HTTP时间字符串，支持多种格式
func ParseHTTPTime(timeStr string) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)

	// 尝试三种HTTP时间格式
	formats := []string{
		HTTPTimeFormat, // RFC 1123
		RFC850Format,   // RFC 850 (废弃但需要支持)
		ASCTimeFormat,  // ANSI C asctime()
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// TimeCache 时间格式化缓存，提高性能
type TimeCache struct {
	mu       sync.RWMutex
	cache    map[int64]string
	lastTime int64
	lastStr  string
}

// NewTimeCache 创建新的时间缓存
func NewTimeCache() *TimeCache {
	return &TimeCache{
		cache: make(map[int64]string),
	}
}

// FormatHTTP 格式化时间为HTTP格式，使用缓存提高性能
func (tc *TimeCache) FormatHTTP(t time.Time) string {
	unix := t.Unix()

	// 快速路径：检查是否是最近使用的时间
	tc.mu.RLock()
	if unix == tc.lastTime {
		result := tc.lastStr
		tc.mu.RUnlock()
		return result
	}

	// 检查缓存
	if cached, exists := tc.cache[unix]; exists {
		tc.mu.RUnlock()
		tc.mu.Lock()
		tc.lastTime = unix
		tc.lastStr = cached
		tc.mu.Unlock()
		return cached
	}
	tc.mu.RUnlock()

	// 格式化新时间
	formatted := FormatHTTPTime(t)

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// 限制缓存大小
	if len(tc.cache) > 1000 {
		// 清空一半缓存
		for k := range tc.cache {
			delete(tc.cache, k)
			if len(tc.cache) <= 500 {
				break
			}
		}
	}

	tc.cache[unix] = formatted
	tc.lastTime = unix
	tc.lastStr = formatted

	return formatted
}

// TimeoutChecker 超时检查器
type TimeoutChecker struct {
	startTime time.Time
	timeout   time.Duration
}

// NewTimeoutChecker 创建新的超时检查器
func NewTimeoutChecker(timeout time.Duration) *TimeoutChecker {
	return &TimeoutChecker{
		startTime: time.Now(),
		timeout:   timeout,
	}
}

// IsExpired 检查是否已超时
func (tc *TimeoutChecker) IsExpired() bool {
	return time.Since(tc.startTime) > tc.timeout
}

// Reset 重置计时器
func (tc *TimeoutChecker) Reset() {
	tc.startTime = time.Now()
}

// Remaining 返回剩余时间
func (tc *TimeoutChecker) Remaining() time.Duration {
	elapsed := time.Since(tc.startTime)
	if elapsed >= tc.timeout {
		return 0
	}
	return tc.timeout - elapsed
}

// ParseDuration 解析持续时间，支持扩展格式
func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// 检查是否为纯数字（默认秒）
	if num, err := strconv.Atoi(s); err == nil {
		return time.Duration(num) * time.Second, nil
	}

	// 处理扩展格式
	if strings.HasSuffix(s, "d") {
		// 天数
		dayStr := strings.TrimSuffix(s, "d")
		if days, err := strconv.Atoi(dayStr); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}

	// 使用标准库解析
	return time.ParseDuration(s)
}

// IsBefore 检查时间t1是否在t2之前
func IsBefore(t1, t2 time.Time) bool {
	return t1.Before(t2)
}

// IsAfter 检查时间t1是否在t2之后
func IsAfter(t1, t2 time.Time) bool {
	return t1.After(t2)
}

// ToUnixTimestamp 转换为Unix时间戳
func ToUnixTimestamp(t time.Time) int64 {
	return t.Unix()
}

// FromUnixTimestamp 从Unix时间戳创建时间
func FromUnixTimestamp(timestamp int64) time.Time {
	return time.Unix(timestamp, 0).UTC()
}

// Age 计算时间的年龄（距离现在的时间）
func Age(t time.Time) time.Duration {
	return time.Since(t)
}

// FormatDuration 格式化持续时间为人类可读格式
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}

	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}

	return fmt.Sprintf("%ds", seconds)
}

// Now 获取当前UTC时间
func Now() time.Time {
	return time.Now().UTC()
}

// StartOfDay 获取指定时间当天的开始时间
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取指定时间当天的结束时间
func EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}

// Truncate 将时间截断到指定精度
func Truncate(t time.Time, d time.Duration) time.Time {
	return t.Truncate(d)
}

// Round 将时间四舍五入到指定精度
func Round(t time.Time, d time.Duration) time.Time {
	return t.Round(d)
}

// SleepUntil 休眠直到指定时间
func SleepUntil(t time.Time) {
	duration := time.Until(t)
	if duration > 0 {
		time.Sleep(duration)
	}
}

// Ticker 高精度定时器
type Ticker struct {
	C        <-chan time.Time
	ticker   *time.Ticker
	interval time.Duration
}

// NewTicker 创建新的定时器
func NewTicker(interval time.Duration) *Ticker {
	ticker := time.NewTicker(interval)
	return &Ticker{
		C:        ticker.C,
		ticker:   ticker,
		interval: interval,
	}
}

// Stop 停止定时器
func (t *Ticker) Stop() {
	if t.ticker != nil {
		t.ticker.Stop()
	}
}

// Reset 重置定时器间隔
func (t *Ticker) Reset(interval time.Duration) {
	if t.ticker != nil {
		t.ticker.Stop()
	}
	t.ticker = time.NewTicker(interval)
	t.C = t.ticker.C
	t.interval = interval
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// NewTimeRange 创建时间范围
func NewTimeRange(start, end time.Time) TimeRange {
	return TimeRange{Start: start, End: end}
}

// Contains 检查时间是否在范围内
func (tr TimeRange) Contains(t time.Time) bool {
	return (t.Equal(tr.Start) || t.After(tr.Start)) &&
		(t.Equal(tr.End) || t.Before(tr.End))
}

// Duration 返回时间范围的持续时间
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// Overlaps 检查两个时间范围是否重叠
func (tr TimeRange) Overlaps(other TimeRange) bool {
	return tr.Start.Before(other.End) && other.Start.Before(tr.End)
}

// 全局时间缓存实例
var globalTimeCache = NewTimeCache()

// CachedHTTPTime 使用全局缓存格式化HTTP时间
func CachedHTTPTime(t time.Time) string {
	return globalTimeCache.FormatHTTP(t)
}

// Timer 高精度计时器
type Timer struct {
	start time.Time
}

// NewTimer 创建新的计时器
func NewTimer() *Timer {
	return &Timer{start: time.Now()}
}

// Elapsed 返回已经过的时间
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.start)
}

// Reset 重置计时器
func (t *Timer) Reset() {
	t.start = time.Now()
}

// ElapsedMillis 返回已经过的毫秒数
func (t *Timer) ElapsedMillis() int64 {
	return t.Elapsed().Milliseconds()
}

// ElapsedMicros 返回已经过的微秒数
func (t *Timer) ElapsedMicros() int64 {
	return t.Elapsed().Microseconds()
}

// ElapsedNanos 返回已经过的纳秒数
func (t *Timer) ElapsedNanos() int64 {
	return t.Elapsed().Nanoseconds()
}
