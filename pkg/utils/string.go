package utils

import (
	"html"
	"net/url"
	"strings"
	"unsafe"
)

// EqualFoldASCII 快速ASCII大小写不敏感比较，专为HTTP头部优化
func EqualFoldASCII(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		c1, c2 := s1[i], s2[i]

		// 转换为小写进行比较
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}

		if c1 != c2 {
			return false
		}
	}

	return true
}

// TrimSpace 快速trim空白字符，优化版本
func TrimSpace(s string) string {
	start := 0
	end := len(s)

	// 从左侧trim
	for start < end && isSpace(s[start]) {
		start++
	}

	// 从右侧trim
	for start < end && isSpace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isSpace 检查是否为空白字符
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// SplitHeader 分割HTTP头部为键值对
func SplitHeader(header string) (key, value string) {
	colon := strings.IndexByte(header, ':')
	if colon == -1 {
		return "", ""
	}

	key = TrimSpace(header[:colon])
	value = TrimSpace(header[colon+1:])
	return key, value
}

// JoinStrings 高效拼接字符串，使用缓冲区
func JoinStrings(sep string, strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	// 计算总长度
	totalLen := 0
	for _, s := range strs {
		totalLen += len(s)
	}
	totalLen += len(sep) * (len(strs) - 1)

	// 使用builder优化性能
	var result strings.Builder
	result.Grow(totalLen)

	result.WriteString(strs[0])
	for i := 1; i < len(strs); i++ {
		result.WriteString(sep)
		result.WriteString(strs[i])
	}

	return result.String()
}

// URLEncode URL编码
func URLEncode(s string) string {
	return url.QueryEscape(s)
}

// URLDecode URL解码
func URLDecode(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s // 解码失败返回原字符串
	}
	return decoded
}

// EscapeHTML HTML转义
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// ContainsAny 检查字符串是否包含任意一个子字符串
func ContainsAny(s string, substrings []string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ReplaceAll 替换所有匹配项
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// IsValidHTTPToken 验证是否为有效的HTTP token
// 根据RFC 7230规范实现
func IsValidHTTPToken(token string) bool {
	if len(token) == 0 {
		return false
	}

	for i := 0; i < len(token); i++ {
		c := token[i]
		// HTTP token字符：VCHAR, except separators
		if !isTokenChar(c) {
			return false
		}
	}

	return true
}

// isTokenChar 检查字符是否为HTTP token有效字符
func isTokenChar(c byte) bool {
	// ASCII printable characters excluding separators
	switch c {
	case '(', ')', '<', '>', '@', ',', ';', ':', '\\', '"', '/', '[', ']', '?', '=', '{', '}', ' ', '\t':
		return false
	default:
		return c > 32 && c < 127 // 可打印ASCII字符
	}
}

// ParseMediaType 解析媒体类型和参数
func ParseMediaType(contentType string) (mediaType string, params map[string]string) {
	params = make(map[string]string)

	// 分割主类型和参数
	parts := strings.Split(contentType, ";")
	if len(parts) == 0 {
		return "", params
	}

	mediaType = TrimSpace(parts[0])

	// 解析参数
	for i := 1; i < len(parts); i++ {
		param := TrimSpace(parts[i])
		eqIdx := strings.IndexByte(param, '=')
		if eqIdx == -1 {
			continue
		}

		key := TrimSpace(param[:eqIdx])
		value := TrimSpace(param[eqIdx+1:])

		// 移除引号
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		params[key] = value
	}

	return mediaType, params
}

// StringToBytes 字符串转字节切片
// 注意：返回的字节切片共享字符串的底层内存，不应修改
func StringToBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	// 使用unsafe进行零拷贝转换，但要小心使用
	return *(*[]byte)(unsafe.Pointer(&struct {
		string
		Cap int
	}{s, len(s)}))
}

// BytesToString 字节切片转字符串
// 注意：不会复制数据，返回的字符串与字节切片共享内存
func BytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&b))
}

// SafeStringToBytes 安全的字符串转字节切片（会复制数据）
func SafeStringToBytes(s string) []byte {
	return []byte(s)
}

// SafeBytesToString 安全的字节切片转字符串（会复制数据）
func SafeBytesToString(b []byte) string {
	return string(b)
}

// FastSplit 快速分割字符串，避免内存分配
func FastSplit(s string, sep byte) []string {
	n := strings.Count(s, string(sep)) + 1
	result := make([]string, 0, n)

	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])

	return result
}

// HasPrefix 检查字符串前缀，优化版本
func HasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// HasSuffix 检查字符串后缀，优化版本
func HasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// ToLowerASCII 快速ASCII转小写
func ToLowerASCII(s string) string {
	// 首先检查是否需要转换
	needConvert := false
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			needConvert = true
			break
		}
	}

	if !needConvert {
		return s
	}

	// 需要转换时创建新字符串
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}

	return string(result)
}

// ToUpperASCII 快速ASCII转大写
func ToUpperASCII(s string) string {
	// 首先检查是否需要转换
	needConvert := false
	for i := 0; i < len(s); i++ {
		if s[i] >= 'a' && s[i] <= 'z' {
			needConvert = true
			break
		}
	}

	if !needConvert {
		return s
	}

	// 需要转换时创建新字符串
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		result[i] = c
	}

	return string(result)
}

// IndexByteN 查找第n个匹配的字节位置
func IndexByteN(s string, c byte, n int) int {
	if n <= 0 {
		return -1
	}

	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			count++
			if count == n {
				return i
			}
		}
	}

	return -1
}

// Clone 复制字符串（Go 1.18+有strings.Clone，这里为兼容性提供）
func Clone(s string) string {
	if len(s) == 0 {
		return ""
	}
	b := make([]byte, len(s))
	copy(b, s)
	return string(b)
}
