package testing

import (
	"fmt"
	"reflect"
	"sync"
)

// MockCall 表示一次模拟调用
type MockCall struct {
	Method       string
	Args         []interface{}
	ReturnValues []interface{}
	CallCount    int
	mutex        sync.RWMutex // 保护CallCount, 避免并发访问
}

// NewMockCall 创建新的模拟调用
func NewMockCall(method string, args []interface{}, returns []interface{}) *MockCall {
	return &MockCall{
		Method:       method,
		Args:         args,
		ReturnValues: returns,
	}
}

// Called 记录调用
func (mc *MockCall) Called() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.CallCount++
}

// GetCallCount 获取调用次数
func (mc *MockCall) GetCallCount() int {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	return mc.CallCount
}

// CallHistory 调用历史记录（不包含锁）
type CallHistory struct {
	Method string
	Args   []interface{}
}

// Mock 模拟对象
type Mock struct {
	calls   map[string]*MockCall
	history []CallHistory
	mutex   sync.RWMutex
}

// NewMock 创建新的模拟对象
func NewMock() *Mock {
	return &Mock{
		calls: make(map[string]*MockCall),
	}
}

// On 设置期望调用
func (m *Mock) On(method string, args ...interface{}) *MockCall {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.getKey(method, args)
	call := NewMockCall(method, args, nil)
	m.calls[key] = call
	return call
}

// Returns 设置返回值
func (mc *MockCall) Returns(values ...interface{}) *MockCall {
	mc.ReturnValues = values
	return mc
}

// Call 调用模拟方法
func (m *Mock) Call(method string, args ...interface{}) []interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := m.getKey(method, args)
	call, exists := m.calls[key]
	if !exists {
		panic(fmt.Sprintf("unexpected call to %s with args %v", method, args))
	}

	call.Called()

	// 记录调用历史
	historyCall := CallHistory{
		Method: method,
		Args:   args,
	}
	m.history = append(m.history, historyCall)

	return call.ReturnValues
}

// AssertCalled 断言方法被调用
func (m *Mock) AssertCalled(method string, args ...interface{}) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := m.getKey(method, args)
	call, exists := m.calls[key]
	return exists && call.GetCallCount() > 0
}

// AssertNotCalled 断言方法未被调用
func (m *Mock) AssertNotCalled(method string, args ...interface{}) bool {
	return !m.AssertCalled(method, args...)
}

// AssertCallCount 断言调用次数
func (m *Mock) AssertCallCount(method string, expectedCount int, args ...interface{}) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := m.getKey(method, args)
	call, exists := m.calls[key]
	if !exists {
		return expectedCount == 0
	}

	return call.GetCallCount() == expectedCount
}

// GetCallHistory 获取调用历史
func (m *Mock) GetCallHistory() []CallHistory {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	history := make([]CallHistory, len(m.history))
	copy(history, m.history)
	return history
}

// Reset 重置模拟对象
func (m *Mock) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, call := range m.calls {
		call.mutex.Lock()
		call.CallCount = 0
		call.mutex.Unlock()
	}
	m.history = nil
}

// Clear 清空所有模拟调用
func (m *Mock) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.calls = make(map[string]*MockCall)
	m.history = nil
}

// getKey 生成调用键
func (m *Mock) getKey(method string, args []interface{}) string {
	return fmt.Sprintf("%s_%v", method, args)
}

// HTTPMock HTTP模拟器
type HTTPMock struct {
	*Mock
	responses map[string]HTTPResponse
}

// HTTPResponse HTTP响应模拟
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Error      error
}

// NewHTTPMock 创建HTTP模拟器
func NewHTTPMock() *HTTPMock {
	return &HTTPMock{
		Mock:      NewMock(),
		responses: make(map[string]HTTPResponse),
	}
}

// MockResponse 模拟HTTP响应
func (h *HTTPMock) MockResponse(url string, response HTTPResponse) {
	h.responses[url] = response
}

// GetResponse 获取模拟响应
func (h *HTTPMock) GetResponse(url string) (HTTPResponse, bool) {
	response, exists := h.responses[url]
	return response, exists
}

// TCPMock TCP连接模拟器
type TCPMock struct {
	*Mock
	connections map[string]*MockConnection
}

// MockConnection 模拟连接
type MockConnection struct {
	Address   string
	Connected bool
	Data      []byte
	Error     error
}

// NewTCPMock 创建TCP模拟器
func NewTCPMock() *TCPMock {
	return &TCPMock{
		Mock:        NewMock(),
		connections: make(map[string]*MockConnection),
	}
}

// MockConnection 模拟TCP连接
func (t *TCPMock) MockConnection(address string, conn *MockConnection) {
	t.connections[address] = conn
}

// GetConnection 获取模拟连接
func (t *TCPMock) GetConnection(address string) (*MockConnection, bool) {
	conn, exists := t.connections[address]
	return conn, exists
}

// ConfigMock 配置模拟器
type ConfigMock struct {
	*Mock
	configs map[string]interface{}
}

// NewConfigMock 创建配置模拟器
func NewConfigMock() *ConfigMock {
	return &ConfigMock{
		Mock:    NewMock(),
		configs: make(map[string]interface{}),
	}
}

// Set 设置配置值
func (c *ConfigMock) Set(key string, value interface{}) {
	c.configs[key] = value
}

// Get 获取配置值
func (c *ConfigMock) Get(key string) (interface{}, bool) {
	value, exists := c.configs[key]
	return value, exists
}

// GetString 获取字符串配置
func (c *ConfigMock) GetString(key string) string {
	if value, exists := c.configs[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt 获取整数配置
func (c *ConfigMock) GetInt(key string) int {
	if value, exists := c.configs[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool 获取布尔配置
func (c *ConfigMock) GetBool(key string) bool {
	if value, exists := c.configs[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// argsMatch 检查参数是否匹配
func argsMatch(expected, actual []interface{}) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i, exp := range expected {
		if !reflect.DeepEqual(exp, actual[i]) {
			return false
		}
	}

	return true
}
