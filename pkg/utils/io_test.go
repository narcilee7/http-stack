package utils

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	testhelper "github.com/narcilee7/http-stack/internal/testing"
)

func TestIOUtils(t *testing.T) {
	h := testhelper.NewHelper(t)

	t.Run("limited_reader", func(t *testing.T) {
		data := "hello world test data"
		reader := strings.NewReader(data)
		limited := NewLimitedReader(reader, 5)

		buf := make([]byte, 10)
		n, err := limited.Read(buf)
		h.AssertEqual(n, 5, "Should read limited bytes")
		h.AssertNil(err, "Should not error on normal read")
		h.AssertEqual(string(buf[:n]), "hello", "Should read correct data")

		// 再次读取应该返回EOF
		n, err = limited.Read(buf)
		h.AssertEqual(n, 0, "Should read 0 bytes after limit")
		h.AssertEqual(err, io.EOF, "Should return EOF after limit")
	})

	t.Run("counting_reader", func(t *testing.T) {
		data := "test counting reader"
		reader := strings.NewReader(data)
		counter := NewCountingReader(reader)

		buf := make([]byte, 4)
		n, err := counter.Read(buf)
		h.AssertEqual(n, 4, "Should read 4 bytes")
		h.AssertNil(err, "Should not error")
		h.AssertEqual(counter.Count(), int64(4), "Counter should show 4 bytes")

		// 读取剩余数据
		remaining, _ := io.ReadAll(counter)
		totalExpected := int64(len(data))
		h.AssertEqual(counter.Count(), totalExpected, "Counter should show total bytes read")
		h.AssertEqual(len(remaining), len(data)-4, "Should read remaining data")
	})

	t.Run("counting_writer", func(t *testing.T) {
		var buf bytes.Buffer
		counter := NewCountingWriter(&buf)

		data1 := []byte("hello")
		n, err := counter.Write(data1)
		h.AssertEqual(n, len(data1), "Should write all bytes")
		h.AssertNil(err, "Should not error")
		h.AssertEqual(counter.Count(), int64(len(data1)), "Counter should show written bytes")

		data2 := []byte(" world")
		counter.Write(data2)
		h.AssertEqual(counter.Count(), int64(len(data1)+len(data2)), "Counter should accumulate")
		h.AssertEqual(buf.String(), "hello world", "Buffer should contain all data")
	})

	t.Run("multi_writer", func(t *testing.T) {
		var buf1, buf2, buf3 bytes.Buffer
		writer := NewMultiWriter(&buf1, &buf2, &buf3)

		data := []byte("test multi writer")
		n, err := writer.Write(data)
		h.AssertEqual(n, len(data), "Should write all bytes")
		h.AssertNil(err, "Should not error")

		// 所有buffer应该包含相同数据
		h.AssertEqual(buf1.String(), string(data), "Buffer1 should contain data")
		h.AssertEqual(buf2.String(), string(data), "Buffer2 should contain data")
		h.AssertEqual(buf3.String(), string(data), "Buffer3 should contain data")
	})

	t.Run("rate_limiter", func(t *testing.T) {
		limiter := NewRateLimiter(1024) // 1KB/s
		data := make([]byte, 512)

		start := time.Now()
		allowed := limiter.Allow(len(data))
		elapsed := time.Since(start)

		h.AssertTrue(allowed, "Should allow first request")
		h.AssertTrue(elapsed < 10*time.Millisecond, "First request should be immediate")

		// 第二次请求应该被限制
		start = time.Now()
		allowed = limiter.Allow(len(data))
		elapsed = time.Since(start)

		h.AssertTrue(allowed, "Should eventually allow second request")
		// 由于是512+512=1024字节，应该在大约1秒内完成
	})

	t.Run("buffered_reader", func(t *testing.T) {
		data := "line1\nline2\nline3\n"
		reader := strings.NewReader(data)
		buffered := NewBufferedReader(reader, 8)

		// 读取第一行
		line, err := buffered.ReadLine()
		h.AssertNil(err, "Should read line without error")
		h.AssertEqual(line, "line1", "Should read first line")

		// 读取第二行
		line, err = buffered.ReadLine()
		h.AssertNil(err, "Should read second line without error")
		h.AssertEqual(line, "line2", "Should read second line")

		// 读取剩余数据
		remaining, err := io.ReadAll(buffered)
		h.AssertNil(err, "Should read remaining data")
		h.AssertEqual(string(remaining), "line3\n", "Should read remaining data")
	})

	t.Run("copy_with_buffer", func(t *testing.T) {
		src := "large data for copy test " + strings.Repeat("x", 1000)
		reader := strings.NewReader(src)
		var writer bytes.Buffer

		buf := GetBuffer()
		defer PutBuffer(buf)

		written, err := CopyWithBuffer(&writer, reader, buf.Bytes()[:0])
		h.AssertNil(err, "Copy should not error")
		h.AssertEqual(written, int64(len(src)), "Should copy all bytes")
		h.AssertEqual(writer.String(), src, "Should copy data correctly")
	})

	t.Run("tee_reader", func(t *testing.T) {
		src := "test tee reader functionality"
		reader := strings.NewReader(src)
		var teeWriter bytes.Buffer
		teeReader := NewTeeReader(reader, &teeWriter)

		// 读取数据
		result, err := io.ReadAll(teeReader)
		h.AssertNil(err, "Should read without error")
		h.AssertEqual(string(result), src, "Should read original data")
		h.AssertEqual(teeWriter.String(), src, "Tee writer should contain copy")
	})

	t.Run("section_reader", func(t *testing.T) {
		data := "0123456789abcdef"
		reader := strings.NewReader(data)
		section := NewSectionReader(reader, 5, 5) // 从位置5读取5个字节

		result, err := io.ReadAll(section)
		h.AssertNil(err, "Should read without error")
		h.AssertEqual(string(result), "56789", "Should read correct section")

		// 检查剩余长度
		h.AssertEqual(section.Size(), int64(5), "Section size should be 5")
	})

	t.Run("discard_writer", func(t *testing.T) {
		writer := DiscardWriter()
		data := []byte("this data will be discarded")

		n, err := writer.Write(data)
		h.AssertEqual(n, len(data), "Should accept all bytes")
		h.AssertNil(err, "Should not error")

		// 多次写入
		for i := 0; i < 100; i++ {
			n, err = writer.Write(data)
			h.AssertEqual(n, len(data), "Should always accept all bytes")
			h.AssertNil(err, "Should never error")
		}
	})
}

func TestIOUtilsPerformance(t *testing.T) {
	h := testhelper.NewHelper(t)

	t.Run("buffered_vs_direct", func(t *testing.T) {
		data := strings.Repeat("test data line\n", 1000)

		// 直接读取
		start := time.Now()
		reader1 := strings.NewReader(data)
		_, err := io.ReadAll(reader1)
		directTime := time.Since(start)
		h.AssertNil(err, "Direct read should not error")

		// 缓冲读取
		start = time.Now()
		reader2 := strings.NewReader(data)
		buffered := NewBufferedReader(reader2, 4096)
		_, err = io.ReadAll(buffered)
		bufferedTime := time.Since(start)
		h.AssertNil(err, "Buffered read should not error")

		// 对于这种场景，差异可能不大，但不应该更慢太多
		h.AssertTrue(bufferedTime < directTime*2, "Buffered read should not be much slower")
	})

	t.Run("copy_buffer_reuse", func(t *testing.T) {
		src := strings.Repeat("x", 10000)

		// 测试池化buffer功能性
		for i := 0; i < 10; i++ {
			reader := strings.NewReader(src)
			var writer bytes.Buffer
			buf := GetBuffer()
			written, err := CopyWithBuffer(&writer, reader, buf.Bytes()[:0])
			h.AssertNil(err, "Copy should not error")
			h.AssertEqual(written, int64(len(src)), "Should copy all bytes")
			PutBuffer(buf)
		}

		// 测试直接buffer功能性
		for i := 0; i < 10; i++ {
			reader := strings.NewReader(src)
			var writer bytes.Buffer
			buf := make([]byte, 32*1024)
			written, err := CopyWithBuffer(&writer, reader, buf)
			h.AssertNil(err, "Copy should not error")
			h.AssertEqual(written, int64(len(src)), "Should copy all bytes")
		}
	})
}

// 基准测试
func BenchmarkIOUtils(b *testing.B) {
	data := strings.Repeat("benchmark test data\n", 1000)

	testhelper.ComparisonBenchmark(b, map[string]func(*testing.B){
		"stdlib_copy": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(data)
				var writer bytes.Buffer
				io.Copy(&writer, reader)
			}
		},
		"buffered_copy": func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(data)
				var writer bytes.Buffer
				buf := GetBuffer()
				CopyWithBuffer(&writer, reader, buf.Bytes()[:0])
				PutBuffer(buf)
			}
		},
	})
}

func BenchmarkCountingReader(b *testing.B) {
	data := strings.Repeat("x", 1024)

	testhelper.MemoryBenchmark(b, "counting_read", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(data)
			counter := NewCountingReader(reader)
			io.ReadAll(counter)
		}
	})
}

func BenchmarkMultiWriter(b *testing.B) {
	data := []byte(strings.Repeat("test", 256))

	testhelper.ProgressiveBenchmark(b, "writers", []int{1, 2, 4, 8}, func(b *testing.B, numWriters int) {
		writers := make([]io.Writer, numWriters)
		for i := 0; i < numWriters; i++ {
			writers[i] = DiscardWriter()
		}
		multiWriter := NewMultiWriter(writers...)

		for i := 0; i < b.N; i++ {
			multiWriter.Write(data)
		}
	})
}

func BenchmarkRateLimiter(b *testing.B) {
	limiter := NewRateLimiter(1024 * 1024) // 1MB/s

	testhelper.MemoryBenchmark(b, "rate_limit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			limiter.Allow(1024) // 1KB请求
		}
	})
}
