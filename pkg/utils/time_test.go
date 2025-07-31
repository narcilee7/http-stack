package utils

import (
	"testing"
	"time"

	testhelper "github.com/narcilee7/http-stack/internal/testing"
)

func TestTimeUtils(t *testing.T) {
	h := testhelper.NewHelper(t)

	t.Run("http_time_format", func(t *testing.T) {
		// HTTP标准时间格式：RFC 7231
		now := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
		formatted := FormatHTTPTime(now)
		expected := "Mon, 15 Jan 2024 10:30:45 GMT"
		h.AssertEqual(formatted, expected, "Should format time in HTTP standard format")

		// 测试解析
		parsed, err := ParseHTTPTime(formatted)
		h.AssertNil(err, "Should parse HTTP time without error")
		h.AssertEqual(parsed.Unix(), now.Unix(), "Parsed time should match original")
	})

	t.Run("parse_various_formats", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected time.Time
		}{
			// RFC 1123 (HTTP标准)
			{"Mon, 15 Jan 2024 10:30:45 GMT", time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)},
			// RFC 850 (废弃但需要支持)
			{"Monday, 15-Jan-24 10:30:45 GMT", time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)},
			// ANSI C asctime()
			{"Mon Jan 15 10:30:45 2024", time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)},
		}

		for _, tc := range testCases {
			parsed, err := ParseHTTPTime(tc.input)
			h.AssertNil(err, "Should parse time format: %s", tc.input)
			h.AssertEqual(parsed.Unix(), tc.expected.Unix(), "Parsed time should match expected for: %s", tc.input)
		}
	})

	t.Run("time_cache", func(t *testing.T) {
		// 时间缓存功能测试
		cache := NewTimeCache()
		now := time.Now().UTC()

		// 第一次获取
		formatted1 := cache.FormatHTTP(now)
		h.AssertTrue(len(formatted1) > 0, "Should return formatted time")

		// 再次获取相同时间，应该使用缓存
		formatted2 := cache.FormatHTTP(now)
		h.AssertEqual(formatted1, formatted2, "Should return same cached result")

		// 不同时间应该返回不同结果
		later := now.Add(time.Second)
		formatted3 := cache.FormatHTTP(later)
		h.AssertNotEqual(formatted1, formatted3, "Different times should have different formats")
	})

	t.Run("timeout_checker", func(t *testing.T) {
		checker := NewTimeoutChecker(100 * time.Millisecond)
		h.AssertFalse(checker.IsExpired(), "Should not be expired immediately")

		// 等待超时
		time.Sleep(150 * time.Millisecond)
		h.AssertTrue(checker.IsExpired(), "Should be expired after timeout")

		// 重置后不应该过期
		checker.Reset()
		h.AssertFalse(checker.IsExpired(), "Should not be expired after reset")
	})

	t.Run("duration_parsing", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected time.Duration
		}{
			{"30s", 30 * time.Second},
			{"5m", 5 * time.Minute},
			{"2h", 2 * time.Hour},
			{"1d", 24 * time.Hour},
			{"30", 30 * time.Second}, // 默认秒
		}

		for _, tc := range testCases {
			duration, err := ParseDuration(tc.input)
			h.AssertNil(err, "Should parse duration: %s", tc.input)
			h.AssertEqual(duration, tc.expected, "Duration should match expected for: %s", tc.input)
		}

		// 无效格式应该返回错误
		_, err := ParseDuration("invalid")
		h.AssertNotNil(err, "Should return error for invalid duration")
	})

	t.Run("time_comparison", func(t *testing.T) {
		base := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
		earlier := base.Add(-time.Hour)
		later := base.Add(time.Hour)

		h.AssertTrue(IsBefore(earlier, base), "Earlier time should be before base")
		h.AssertFalse(IsBefore(later, base), "Later time should not be before base")
		h.AssertFalse(IsBefore(base, base), "Same time should not be before itself")

		h.AssertTrue(IsAfter(later, base), "Later time should be after base")
		h.AssertFalse(IsAfter(earlier, base), "Earlier time should not be after base")
		h.AssertFalse(IsAfter(base, base), "Same time should not be after itself")
	})

	t.Run("unix_timestamp", func(t *testing.T) {
		now := time.Now().UTC()
		timestamp := ToUnixTimestamp(now)
		recovered := FromUnixTimestamp(timestamp)

		h.AssertEqual(now.Unix(), recovered.Unix(), "Unix timestamp conversion should be accurate")
	})

	t.Run("age_calculation", func(t *testing.T) {
		past := time.Now().Add(-5 * time.Minute)
		age := Age(past)

		// 应该大约是5分钟
		h.AssertTrue(age >= 4*time.Minute && age <= 6*time.Minute, "Age should be approximately 5 minutes")
	})

	t.Run("format_duration", func(t *testing.T) {
		testCases := []struct {
			duration time.Duration
			expected string
		}{
			{30 * time.Second, "30s"},
			{5 * time.Minute, "5m0s"},
			{2 * time.Hour, "2h0m0s"},
			{25 * time.Hour, "25h0m0s"},
		}

		for _, tc := range testCases {
			formatted := FormatDuration(tc.duration)
			h.AssertEqual(formatted, tc.expected, "Duration format should match expected")
		}
	})
}

func TestTimeUtilsPerformance(t *testing.T) {
	h := testhelper.NewHelper(t)

	t.Run("time_cache_performance", func(t *testing.T) {
		cache := NewTimeCache()
		now := time.Now().UTC()

		// 预热缓存
		cache.FormatHTTP(now)

		// 测试缓存命中性能
		start := time.Now()
		for i := 0; i < 1000; i++ {
			cache.FormatHTTP(now)
		}
		elapsed := time.Since(start)

		// 1000次缓存命中应该很快
		h.AssertTrue(elapsed < 10*time.Millisecond, "Cache hits should be very fast")
	})

	t.Run("timeout_check_performance", func(t *testing.T) {
		checker := NewTimeoutChecker(time.Hour) // 不会过期

		start := time.Now()
		for i := 0; i < 10000; i++ {
			checker.IsExpired()
		}
		elapsed := time.Since(start)

		// 10000次超时检查应该很快
		h.AssertTrue(elapsed < 50*time.Millisecond, "Timeout checks should be very fast")
	})
}

// 基准测试
func BenchmarkTimeUtils(b *testing.B) {
	now := time.Now().UTC()

	testhelper.ComparisonBenchmark(b, map[string]func(*testing.B){
		"stdlib_format": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = now.Format(time.RFC1123)
			}
		},
		"custom_format": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = FormatHTTPTime(now)
			}
		},
	})
}

func BenchmarkTimeCache(b *testing.B) {
	cache := NewTimeCache()
	now := time.Now().UTC()

	testhelper.ComparisonBenchmark(b, map[string]func(*testing.B){
		"no_cache": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = FormatHTTPTime(now)
			}
		},
		"with_cache": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = cache.FormatHTTP(now)
			}
		},
	})
}

func BenchmarkTimeParsing(b *testing.B) {
	timeStr := "Mon, 15 Jan 2024 10:30:45 GMT"

	testhelper.ComparisonBenchmark(b, map[string]func(*testing.B){
		"stdlib_parse": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = time.Parse(time.RFC1123, timeStr)
			}
		},
		"custom_parse": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ParseHTTPTime(timeStr)
			}
		},
	})
}

func BenchmarkTimeoutCheck(b *testing.B) {
	checker := NewTimeoutChecker(time.Hour)

	testhelper.MemoryBenchmark(b, "timeout_check", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = checker.IsExpired()
		}
	})
}
