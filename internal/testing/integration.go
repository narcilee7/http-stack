package testing

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

// IntegrationTest 集成测试结构
type IntegrationTest struct {
	Name      string
	Setup     func() error
	Test      func(*testing.T) error
	Teardown  func() error
	Timeout   time.Duration
	SkipShort bool
}

// Run 运行集成测试
func (it *IntegrationTest) Run(t *testing.T) {
	if testing.Short() && it.SkipShort {
		// 跳过集成测试
		t.Skip("Skipping integration test in short mode")
	}
	t.Run(it.Name, func(t *testing.T) {
		// 设置超时
		if it.Timeout > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), it.Timeout)
			defer cancel()

			done := make(chan bool, 1)

			go func() {
				defer func() { done <- true }()
				it.runTest(t)
			}()

			select {
			case <-done:
			// 测试完成
			case <-ctx.Done():
				t.Fatal("Integration test timed out")
			}
		} else {
			it.runTest(t)
		}
	})
}

// runTest 实际运行测试
func (it *IntegrationTest) runTest(t *testing.T) {
	// Setup阶段
	if it.Setup != nil {
		if err := it.Setup(); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// 确保Teardown执行
	defer func() {
		if it.Teardown != nil {
			if err := it.Teardown(); err != nil {
				t.Errorf("Teardown failed: %v", err)
			}
		}
	}()

	// 运行测试
	if err := it.Test(t); err != nil {
		t.Fatalf("Test failed: %v", err)
	}
}

// TestSuite 测试套件
type TestSuite struct {
	Name          string
	SetupSuite    func() error
	TeardownSuite func() error
	SetupTest     func() error
	TeardownTest  func() error
	Tests         []IntegrationTest
	Parallel      bool
}

// Run 运行测试套件
func (ts *TestSuite) Run(t *testing.T) {
	t.Run(ts.Name, func(t *testing.T) {
		// Suite级别的Setup
		if ts.SetupSuite != nil {
			if err := ts.SetupSuite(); err != nil {
				t.Fatalf("Suite setup failed: %v", err)
			}
		}

		// 确保Suite级别的Teardown执行
		defer func() {
			if ts.TeardownSuite != nil {
				if err := ts.TeardownSuite(); err != nil {
					t.Errorf("Suite teardown failed: %v", err)
				}
			}
		}()

		// 运行测试
		if ts.Parallel {
			ts.runParallel(t)
		} else {
			ts.runSequential(t)
		}
	})
}

// runSequential 顺序运行测试
func (ts *TestSuite) runSequential(t *testing.T) {
	for _, test := range ts.Tests {
		ts.runSingleTest(t, test)
	}
}

// runParallel 并行运行测试
func (ts *TestSuite) runParallel(t *testing.T) {
	for _, test := range ts.Tests {
		test := test // 捕获循环变量
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			ts.runSingleTest(t, test)
		})
	}
}

// runSingleTest 运行单个测试
func (ts *TestSuite) runSingleTest(t *testing.T, test IntegrationTest) {
	// Test级别的Setup
	if ts.SetupTest != nil {
		if err := ts.SetupTest(); err != nil {
			t.Fatalf("Test setup failed: %v", err)
		}
	}

	// 确保Test级别的Teardown执行
	defer func() {
		if ts.TeardownTest != nil {
			if err := ts.TeardownTest(); err != nil {
				t.Errorf("Test teardown failed: %v", err)
			}
		}
	}()

	// 运行测试
	test.Run(t)
}

// ServerTestHelper 服务器测试辅助器
type ServerTestHelper struct {
	Address string
	server  net.Listener
	Port    int
	running bool
	mutex   sync.RWMutex
}

// NewServerTestHelper 创建服务器测试辅助器
func NewServerTestHelper() *ServerTestHelper {
	return &ServerTestHelper{}
}

// Start 启动测试服务器
func (sth *ServerTestHelper) Start() error {
	sth.mutex.Lock()
	defer sth.mutex.Unlock()

	if sth.running {
		return fmt.Errorf("server already running")
	}

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	sth.server = listener
	sth.Address = listener.Addr().String()
	_, portStr, err := net.SplitHostPort(sth.Address)
	if err != nil {
		return fmt.Errorf("failed to parse address: %v", err)
	}

	fmt.Sscanf(portStr, "%d", &sth.Port)
	sth.running = true

	return nil
}

// Stop 停止测试服务器
func (sth *ServerTestHelper) Stop() error {
	sth.mutex.Lock()
	defer sth.mutex.Unlock()

	if !sth.running {
		return nil
	}

	err := sth.server.Close()
	sth.running = false
	return err
}

// IsRunning 检查服务器是否运行
func (sth *ServerTestHelper) IsRunning() bool {
	sth.mutex.RLock()
	defer sth.mutex.RUnlock()
	return sth.running
}

// GetAddress 获取服务器地址
func (sth *ServerTestHelper) GetAddress() string {
	sth.mutex.RLock()
	defer sth.mutex.RUnlock()
	return sth.Address
}

// GetPort 获取服务器端口
func (sth *ServerTestHelper) GetPort() int {
	sth.mutex.RLock()
	defer sth.mutex.RUnlock()
	return sth.Port
}

// ClientTestHelper 客户端测试辅助器
type ClientTestHelper struct {
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
}

// NewClientTestHelper 创建客户端测试辅助器
func NewClientTestHelper() *ClientTestHelper {
	return &ClientTestHelper{
		Timeout:    30 * time.Second,
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	}
}

// WaitForServer 等待服务器可用
func (cth *ClientTestHelper) WaitForServer(address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), cth.Timeout)
	defer cancel()

	for i := 0; i < cth.RetryCount; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server %s", address)
		default:
		}

		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err == nil {
			conn.Close()
			return nil
		}

		if i < cth.RetryCount-1 {
			time.Sleep(cth.RetryDelay)
		}
	}

	return fmt.Errorf("server %s is not available after %d retries", address, cth.RetryCount)
}

// RetryOperation 重试操作
func (cth *ClientTestHelper) RetryOperation(operation func() error) error {
	var lastError error

	for i := 0; i < cth.RetryCount; i++ {
		if err := operation(); err != nil {
			lastError = err
			if i < cth.RetryCount-1 {
				time.Sleep(cth.RetryDelay)
			}
			continue
		}
		return nil
	}

	return fmt.Errorf("operation failed after %d retries, last error: %v", cth.RetryCount, lastError)
}

// PortManager 端口管理器
type PortManager struct {
	usedPorts map[int]bool
	mutex     sync.Mutex
}

// NewPortManager 创建端口管理器
func NewPortManager() *PortManager {
	return &PortManager{
		usedPorts: make(map[int]bool),
	}
}

// GetFreePort 获取空闲端口
func (pm *PortManager) GetFreePort() (int, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port

	if pm.usedPorts[port] {
		return pm.GetFreePort() // 递归获取另一个端口
	}

	pm.usedPorts[port] = true
	return port, nil
}

// ReleasePort 释放端口
func (pm *PortManager) ReleasePort(port int) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	delete(pm.usedPorts, port)
}

// GetUsedPorts 获取已使用的端口
func (pm *PortManager) GetUsedPorts() []int {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	ports := make([]int, 0, len(pm.usedPorts))
	for port := range pm.usedPorts {
		ports = append(ports, port)
	}
	return ports
}

// TestEnvironment 测试环境
type TestEnvironment struct {
	Name        string
	Components  map[string]interface{}
	PortManager *PortManager
	mutex       sync.RWMutex
}

// NewTestEnvironment 创建测试环境
func NewTestEnvironment(name string) *TestEnvironment {
	return &TestEnvironment{
		Name:        name,
		Components:  make(map[string]interface{}),
		PortManager: NewPortManager(),
	}
}

// AddComponent 添加组件
func (te *TestEnvironment) AddComponent(name string, component interface{}) {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	te.Components[name] = component
}

// GetComponent 获取组件
func (te *TestEnvironment) GetComponent(name string) (interface{}, bool) {
	te.mutex.RLock()
	defer te.mutex.RUnlock()
	component, exists := te.Components[name]
	return component, exists
}

// RemoveComponent 移除组件
func (te *TestEnvironment) RemoveComponent(name string) {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	delete(te.Components, name)
}

// Cleanup 清理测试环境
func (te *TestEnvironment) Cleanup() {
	te.mutex.Lock()
	defer te.mutex.Unlock()

	// 清理所有组件
	for name, component := range te.Components {
		if cleanup, ok := component.(interface{ Cleanup() error }); ok {
			cleanup.Cleanup()
		}
		delete(te.Components, name)
	}
}
