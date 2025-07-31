package utils

import (
	"bufio"
	"io"
	"sync"
	"time"
)

// LimitedReader 限制读取字节数的Reader
type LimitedReader struct {
	r io.Reader
	n int64
}

// NewLimitedReader 创建限制读取字节数的Reader
func NewLimitedReader(r io.Reader, n int64) *LimitedReader {
	return &LimitedReader{r: r, n: n}
}

// Read 实现io.Reader接口
func (lr *LimitedReader) Read(p []byte) (int, error) {
	if lr.n <= 0 {
		return 0, io.EOF
	}

	if int64(len(p)) > lr.n {
		p = p[0:lr.n]
	}

	n, err := lr.r.Read(p)
	lr.n -= int64(n)
	return n, err
}

// CountingReader 计数Reader，统计读取的字节数
type CountingReader struct {
	r     io.Reader
	count int64
	mutex sync.RWMutex
}

// NewCountingReader 创建计数Reader
func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{r: r}
}

// Read 实现io.Reader接口
func (cr *CountingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.mutex.Lock()
	cr.count += int64(n)
	cr.mutex.Unlock()
	return n, err
}

// Count 返回已读取的字节数
func (cr *CountingReader) Count() int64 {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()
	return cr.count
}

// Reset 重置计数器
func (cr *CountingReader) Reset() {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	cr.count = 0
}

// CountingWriter 计数Writer，统计写入的字节数
type CountingWriter struct {
	w     io.Writer
	count int64
	mutex sync.RWMutex
}

// NewCountingWriter 创建计数Writer
func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{w: w}
}

// Write 实现io.Writer接口
func (cw *CountingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.mutex.Lock()
	cw.count += int64(n)
	cw.mutex.Unlock()
	return n, err
}

// Count 返回已写入的字节数
func (cw *CountingWriter) Count() int64 {
	cw.mutex.RLock()
	defer cw.mutex.RUnlock()
	return cw.count
}

// Reset 重置计数器
func (cw *CountingWriter) Reset() {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()
	cw.count = 0
}

// MultiWriter 多重Writer，同时写入多个目标
type MultiWriter struct {
	writers []io.Writer
}

// NewMultiWriter 创建多重Writer
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{writers: writers}
}

// Write 实现io.Writer接口
func (mw *MultiWriter) Write(p []byte) (int, error) {
	for _, w := range mw.writers {
		n, err := w.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(p) {
			return n, io.ErrShortWrite
		}
	}
	return len(p), nil
}

// RateLimiter 速率限制器
type RateLimiter struct {
	rate     int64     // 每秒字节数
	tokens   int64     // 当前令牌数
	lastTime time.Time // 上次更新时间
	mutex    sync.Mutex
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(bytesPerSecond int64) *RateLimiter {
	return &RateLimiter{
		rate:     bytesPerSecond,
		tokens:   bytesPerSecond,
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许指定字节数的操作
func (rl *RateLimiter) Allow(bytes int) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime)

	// 添加新令牌
	newTokens := int64(elapsed.Seconds() * float64(rl.rate))
	rl.tokens += newTokens
	if rl.tokens > rl.rate {
		rl.tokens = rl.rate
	}
	rl.lastTime = now

	// 检查是否有足够令牌
	if rl.tokens >= int64(bytes) {
		rl.tokens -= int64(bytes)
		return true
	}

	// 等待获取足够令牌
	waitTime := time.Duration(float64(int64(bytes)-rl.tokens) / float64(rl.rate) * float64(time.Second))
	if waitTime > 0 {
		time.Sleep(waitTime)
		rl.tokens = 0
	}

	return true
}

// BufferedReader 缓冲Reader
type BufferedReader struct {
	*bufio.Reader
}

// NewBufferedReader 创建缓冲Reader
func NewBufferedReader(r io.Reader, size int) *BufferedReader {
	return &BufferedReader{
		Reader: bufio.NewReaderSize(r, size),
	}
}

// ReadLine 读取一行（去除换行符）
func (br *BufferedReader) ReadLine() (string, error) {
	line, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	// 去除换行符
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	return line, err
}

// CopyWithBuffer 使用指定缓冲区进行复制
func CopyWithBuffer(dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if len(buf) == 0 {
		buf = make([]byte, 32*1024) // 默认32KB
	}

	var written int64
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = io.ErrShortWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return written, er
			}
			break
		}
	}
	return written, nil
}

// TeeReader 分流Reader，同时写入到Writer
type TeeReader struct {
	r io.Reader
	w io.Writer
}

// NewTeeReader 创建分流Reader
func NewTeeReader(r io.Reader, w io.Writer) *TeeReader {
	return &TeeReader{r: r, w: w}
}

// Read 实现io.Reader接口
func (tr *TeeReader) Read(p []byte) (int, error) {
	n, err := tr.r.Read(p)
	if n > 0 {
		if n, err := tr.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return n, err
}

// SectionReader 区段Reader，读取指定范围的数据
type SectionReader struct {
	r     io.ReaderAt
	off   int64
	limit int64
	n     int64
}

// NewSectionReader 创建区段Reader
func NewSectionReader(r io.ReaderAt, off, n int64) *SectionReader {
	return &SectionReader{r: r, off: off, limit: off + n, n: n}
}

// Read 实现io.Reader接口
func (sr *SectionReader) Read(p []byte) (int, error) {
	if sr.off >= sr.limit {
		return 0, io.EOF
	}
	if max := sr.limit - sr.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err := sr.r.ReadAt(p, sr.off)
	sr.off += int64(n)
	return n, err
}

// Size 返回区段大小
func (sr *SectionReader) Size() int64 {
	return sr.n
}

// discardWriter 丢弃Writer，不保存任何数据
type discardWriter struct{}

// DiscardWriter 返回丢弃所有数据的Writer
func DiscardWriter() io.Writer {
	return discardWriter{}
}

// Write 实现io.Writer接口，丢弃所有数据
func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// NopCloser 将Reader包装为ReadCloser，Close操作为空
func NopCloser(r io.Reader) io.ReadCloser {
	return nopCloser{r}
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

// CopyN 复制指定字节数
func CopyN(dst io.Writer, src io.Reader, n int64) (int64, error) {
	written, err := io.CopyN(dst, src, n)
	return written, err
}

// ReadFull 完整读取指定字节数
func ReadFull(r io.Reader, buf []byte) (int, error) {
	return io.ReadFull(r, buf)
}

// ReadAtLeast 至少读取指定字节数
func ReadAtLeast(r io.Reader, buf []byte, min int) (int, error) {
	return io.ReadAtLeast(r, buf, min)
}

// WriteString 写入字符串
func WriteString(w io.Writer, s string) (int, error) {
	return io.WriteString(w, s)
}

// Pipe 创建管道
func Pipe() (io.Reader, io.Writer) {
	return io.Pipe()
}

// LimitWriter 限制写入字节数的Writer
type LimitWriter struct {
	w io.Writer
	n int64
}

// NewLimitWriter 创建限制写入字节数的Writer
func NewLimitWriter(w io.Writer, n int64) *LimitWriter {
	return &LimitWriter{w: w, n: n}
}

// Write 实现io.Writer接口
func (lw *LimitWriter) Write(p []byte) (int, error) {
	if lw.n <= 0 {
		return 0, io.ErrShortWrite
	}

	if int64(len(p)) > lw.n {
		p = p[0:lw.n]
	}

	n, err := lw.w.Write(p)
	lw.n -= int64(n)
	return n, err
}

// TimeoutReader 带超时的Reader
type TimeoutReader struct {
	r       io.Reader
	timeout time.Duration
}

// NewTimeoutReader 创建带超时的Reader
func NewTimeoutReader(r io.Reader, timeout time.Duration) *TimeoutReader {
	return &TimeoutReader{r: r, timeout: timeout}
}

// Read 实现io.Reader接口
func (tr *TimeoutReader) Read(p []byte) (int, error) {
	type result struct {
		n   int
		err error
	}

	ch := make(chan result, 1)
	go func() {
		n, err := tr.r.Read(p)
		ch <- result{n, err}
	}()

	select {
	case res := <-ch:
		return res.n, res.err
	case <-time.After(tr.timeout):
		return 0, io.ErrUnexpectedEOF
	}
}

// ChunkedReader 分块Reader，用于HTTP chunked编码
type ChunkedReader struct {
	r      io.Reader
	buffer []byte
	offset int
	size   int
}

// NewChunkedReader 创建分块Reader
func NewChunkedReader(r io.Reader) *ChunkedReader {
	return &ChunkedReader{
		r:      r,
		buffer: make([]byte, 4096),
	}
}

// Read 实现io.Reader接口
func (cr *ChunkedReader) Read(p []byte) (int, error) {
	// 简化实现，实际HTTP chunked解码会更复杂
	return cr.r.Read(p)
}

// ReadCloserWrapper 将Reader和Closer组合成ReadCloser
type ReadCloserWrapper struct {
	io.Reader
	io.Closer
}

// NewReadCloserWrapper 创建ReadCloser包装器
func NewReadCloserWrapper(r io.Reader, c io.Closer) io.ReadCloser {
	return &ReadCloserWrapper{Reader: r, Closer: c}
}

// WriteCloserWrapper 将Writer和Closer组合成WriteCloser
type WriteCloserWrapper struct {
	io.Writer
	io.Closer
}

// NewWriteCloserWrapper 创建WriteCloser包装器
func NewWriteCloserWrapper(w io.Writer, c io.Closer) io.WriteCloser {
	return &WriteCloserWrapper{Writer: w, Closer: c}
}
