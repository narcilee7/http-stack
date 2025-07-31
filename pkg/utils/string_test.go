package utils

import (
	"strings"
	"testing"

	testhelper "github.com/narcilee7/http-stack/internal/testing"
)

func TestStringUtils(t *testing.T) {
	h := testhelper.NewHelper(t)

	t.Run("case_insensitive_compare", func(t *testing.T) {
		h.AssertTrue(EqualFoldASCII("Content-Type", "content-type"), "Should be case insensitive")
		h.AssertTrue(EqualFoldASCII("HOST", "host"), "Should be case insensitive")
		h.AssertFalse(EqualFoldASCII("Content-Type", "content-length"), "Different strings should not match")
		h.AssertTrue(EqualFoldASCII("", ""), "Empty strings should match")
	})

	t.Run("fast_trim", func(t *testing.T) {
		h.AssertEqual(TrimSpace("  hello world  "), "hello world", "Should trim spaces")
		h.AssertEqual(TrimSpace("\t\n hello \r\n"), "hello", "Should trim all whitespace")
		h.AssertEqual(TrimSpace(""), "", "Empty string should remain empty")
		h.AssertEqual(TrimSpace("no-spaces"), "no-spaces", "No spaces should remain unchanged")
	})

	t.Run("split_header", func(t *testing.T) {
		key, value := SplitHeader("Content-Type: application/json")
		h.AssertEqual(key, "Content-Type", "Should extract header key")
		h.AssertEqual(value, "application/json", "Should extract header value")

		key, value = SplitHeader("Host: example.com:8080")
		h.AssertEqual(key, "Host", "Should extract host key")
		h.AssertEqual(value, "example.com:8080", "Should extract host value")

		key, value = SplitHeader("Invalid header without colon")
		h.AssertEqual(key, "", "Invalid header should return empty key")
		h.AssertEqual(value, "", "Invalid header should return empty value")
	})

	t.Run("join_strings", func(t *testing.T) {
		result := JoinStrings("; ", "boundary=", "form-data", "name=\"file\"")
		expected := "boundary=; form-data; name=\"file\""
		h.AssertEqual(result, expected, "Should join strings with separator")

		result = JoinStrings("", "a", "b", "c")
		h.AssertEqual(result, "abc", "Should join without separator")

		result = JoinStrings(",", "single")
		h.AssertEqual(result, "single", "Single string should not add separator")
	})

	t.Run("url_encode_decode", func(t *testing.T) {
		original := "hello world!@#$%^&*()"
		encoded := URLEncode(original)
		decoded := URLDecode(encoded)
		h.AssertEqual(decoded, original, "Should encode and decode correctly")

		h.AssertEqual(URLEncode("simple"), "simple", "Simple strings should not change")
		h.AssertEqual(URLDecode("hello%20world"), "hello world", "Should decode percent encoding")
	})

	t.Run("escape_html", func(t *testing.T) {
		h.AssertEqual(EscapeHTML("<script>alert('xss')</script>"), "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;", "Should escape HTML")
		h.AssertEqual(EscapeHTML("safe text"), "safe text", "Safe text should not change")
		h.AssertEqual(EscapeHTML(""), "", "Empty string should remain empty")
	})

	t.Run("contains_any", func(t *testing.T) {
		h.AssertTrue(ContainsAny("hello world", []string{"world", "foo"}), "Should find existing substring")
		h.AssertFalse(ContainsAny("hello world", []string{"foo", "bar"}), "Should not find non-existing substrings")
		h.AssertFalse(ContainsAny("", []string{"test"}), "Empty string should not contain anything")
	})

	t.Run("replace_all", func(t *testing.T) {
		result := ReplaceAll("hello world hello", "hello", "hi")
		h.AssertEqual(result, "hi world hi", "Should replace all occurrences")

		result = ReplaceAll("no match", "foo", "bar")
		h.AssertEqual(result, "no match", "No matches should remain unchanged")
	})

	t.Run("validate_http_token", func(t *testing.T) {
		h.AssertTrue(IsValidHTTPToken("Content-Type"), "Should validate valid HTTP token")
		h.AssertTrue(IsValidHTTPToken("X-Custom-Header"), "Should validate custom header")
		h.AssertFalse(IsValidHTTPToken("Invalid Token"), "Should reject tokens with spaces")
		h.AssertFalse(IsValidHTTPToken(""), "Should reject empty token")
		h.AssertFalse(IsValidHTTPToken("Invalid\x00Token"), "Should reject tokens with control characters")
	})

	t.Run("parse_media_type", func(t *testing.T) {
		mediaType, params := ParseMediaType("text/html; charset=utf-8")
		h.AssertEqual(mediaType, "text/html", "Should extract media type")
		h.AssertEqual(params["charset"], "utf-8", "Should extract charset parameter")

		mediaType, params = ParseMediaType("application/json")
		h.AssertEqual(mediaType, "application/json", "Should handle type without parameters")
		h.AssertEqual(len(params), 0, "Should have no parameters")
	})

	t.Run("byte_conversion", func(t *testing.T) {
		s := "hello world"
		b := StringToBytes(s)
		s2 := BytesToString(b)
		h.AssertEqual(s, s2, "Should convert string to bytes and back")
		h.AssertEqual(len(b), len(s), "Byte slice length should match string length")
	})
}

func TestStringUtilsPerformance(t *testing.T) {
	h := testhelper.NewHelper(t)

	t.Run("zero_copy_conversion", func(t *testing.T) {
		original := "test string for zero copy"
		bytes := StringToBytes(original)
		converted := BytesToString(bytes)

		h.AssertEqual(original, converted, "Zero copy conversion should preserve content")
		h.AssertEqual(len(bytes), len(original), "Byte slice length should match original")

		// 注意：不能修改从StringToBytes得到的字节切片，因为它指向字符串的只读内存
		// 这里只测试转换的正确性，不测试修改
	})
}

// 基准测试
func BenchmarkStringUtils(b *testing.B) {
	testhelper.ComparisonBenchmark(b, map[string]func(*testing.B){
		"stdlib_equal_fold": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				strings.EqualFold("Content-Type", "content-type")
			}
		},
		"fast_equal_fold": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				EqualFoldASCII("Content-Type", "content-type")
			}
		},
	})
}

func BenchmarkStringConversion(b *testing.B) {
	testStr := "benchmark test string for conversion performance"

	testhelper.ComparisonBenchmark(b, map[string]func(*testing.B){
		"safe_conversion": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bytes := []byte(testStr)
				_ = string(bytes)
			}
		},
		"direct_access": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// 直接使用字符串，避免unsafe操作的基准测试问题
				_ = len(testStr)
			}
		},
	})
}

func BenchmarkStringOperations(b *testing.B) {
	testhelper.ProgressiveBenchmark(b, "string_lengths", []int{10, 100, 1000}, func(b *testing.B, size int) {
		testStr := strings.Repeat("a", size)
		for i := 0; i < b.N; i++ {
			_ = TrimSpace("  " + testStr + "  ")
		}
	})
}
