# HTTP协议栈TDD开发策略

## TDD开发流程

### 经典TDD循环：红-绿-重构

1. **红色阶段** - 编写失败的测试
2. **绿色阶段** - 编写最少代码让测试通过
3. **重构阶段** - 优化代码结构，保持测试通过

## 分层TDD策略

### 1. 自底向上的测试策略

```
应用层测试 (HTTP Server/Client)
    ↑
协议层测试 (HTTP/1.1, HTTP/2)
    ↑
传输层测试 (TCP, TLS)
    ↑
工具层测试 (Utils, Buffer Pool)
```

### 2. 开发优先级

**第一阶段：基础工具层**
- 缓冲区管理
- 字符串工具
- 时间工具
- 对象池

**第二阶段：HTTP消息解析**
- HTTP方法解析
- 状态码处理
- 头部解析
- URL解析

**第三阶段：HTTP/1.1协议**
- 请求解析
- 响应生成
- 分块传输
- Keep-Alive

**第四阶段：TCP连接管理**
- 连接池
- 监听器
- 缓冲区

**第五阶段：HTTP服务器**
- 基础服务器
- 路由器
- 中间件

**第六阶段：HTTP客户端**
- 基础客户端
- 连接复用
- 重试机制

## 测试文件结构

```
test/
├── unit/                           # 单元测试
│   ├── utils/                      # 工具层测试
│   │   ├── buffer_pool_test.go
│   │   ├── string_test.go
│   │   └── time_test.go
│   ├── http/                       # HTTP层测试
│   │   ├── message/                # 消息测试
│   │   │   ├── request_test.go
│   │   │   ├── response_test.go
│   │   │   ├── header_test.go
│   │   │   └── cookie_test.go
│   │   ├── protocol/               # 协议测试
│   │   │   ├── http1/
│   │   │   │   ├── parser_test.go
│   │   │   │   ├── writer_test.go
│   │   │   │   └── chunked_test.go
│   │   │   ├── http2/
│   │   │   │   ├── frame_test.go
│   │   │   │   ├── stream_test.go
│   │   │   │   └── hpack_test.go
│   │   │   └── common/
│   │   │       ├── method_test.go
│   │   │       ├── status_test.go
│   │   │       └── url_test.go
│   │   ├── server/                 # 服务器测试
│   │   │   ├── server_test.go
│   │   │   ├── router_test.go
│   │   │   ├── handler_test.go
│   │   │   └── middleware_test.go
│   │   └── client/                 # 客户端测试
│   │       ├── client_test.go
│   │       ├── transport_test.go
│   │       └── retry_test.go
│   ├── tcp/                        # TCP层测试
│   │   ├── listener_test.go
│   │   ├── connection_test.go
│   │   └── buffer_test.go
│   └── tls/                        # TLS层测试
│       ├── config_test.go
│       └── certificate_test.go
├── integration/                    # 集成测试
│   ├── http1_integration_test.go
│   ├── http2_integration_test.go
│   ├── tls_integration_test.go
│   └── end_to_end_test.go
├── performance/                    # 性能测试
│   ├── benchmark_test.go
│   ├── memory_test.go
│   └── concurrency_test.go
└── testdata/                       # 测试数据
    ├── requests/                   # 测试请求
    ├── responses/                  # 测试响应
    └── certificates/               # 测试证书
```

## TDD开发示例

### 示例1：HTTP方法解析器

**第1步：编写失败测试**
```go
// pkg/http/protocol/common/method_test.go
func TestParseMethod(t *testing.T) {
    tests := []struct {
        input    string
        expected Method
        wantErr  bool
    }{
        {"GET", MethodGet, false},
        {"POST", MethodPost, false},
        {"PUT", MethodPut, false},
        {"INVALID", MethodUnknown, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            got, err := ParseMethod(tt.input)
            if tt.wantErr && err == nil {
                t.Error("expected error but got none")
            }
            if !tt.wantErr && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

**第2步：实现最少代码**
```go
// pkg/http/protocol/common/method.go
type Method int

const (
    MethodUnknown Method = iota
    MethodGet
    MethodPost
    MethodPut
)

func ParseMethod(s string) (Method, error) {
    switch s {
    case "GET":
        return MethodGet, nil
    case "POST":
        return MethodPost, nil
    case "PUT":
        return MethodPut, nil
    default:
        return MethodUnknown, errors.New("unknown method")
    }
}
```

**第3步：重构优化**
```go
// 添加更多方法和优化
var methodMap = map[string]Method{
    "GET":     MethodGet,
    "POST":    MethodPost,
    "PUT":     MethodPut,
    "DELETE":  MethodDelete,
    "HEAD":    MethodHead,
    "OPTIONS": MethodOptions,
    "PATCH":   MethodPatch,
}

func ParseMethod(s string) (Method, error) {
    if method, ok := methodMap[s]; ok {
        return method, nil
    }
    return MethodUnknown, fmt.Errorf("unknown HTTP method: %s", s)
}
```

### 示例2：HTTP请求解析器

**第1步：编写失败测试**
```go
// pkg/http/protocol/http1/parser_test.go
func TestParseRequest(t *testing.T) {
    rawRequest := "GET /path?query=value HTTP/1.1\r\n" +
                 "Host: example.com\r\n" +
                 "User-Agent: test-agent\r\n" +
                 "\r\n"
    
    req, err := ParseRequest([]byte(rawRequest))
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if req.Method != MethodGet {
        t.Errorf("expected GET, got %v", req.Method)
    }
    
    if req.Path != "/path" {
        t.Errorf("expected /path, got %s", req.Path)
    }
    
    if req.Query != "query=value" {
        t.Errorf("expected query=value, got %s", req.Query)
    }
    
    if req.Headers.Get("Host") != "example.com" {
        t.Errorf("expected example.com, got %s", req.Headers.Get("Host"))
    }
}
```

### 示例3：HTTP服务器基础功能

**第1步：编写失败测试**
```go
// pkg/http/server/server_test.go
func TestServerBasicHandling(t *testing.T) {
    server := NewServer()
    
    // 注册处理器
    server.RegisterHandler("/hello", func(w ResponseWriter, r *Request) {
        w.WriteHeader(200)
        w.Write([]byte("Hello, World!"))
    })
    
    // 启动测试服务器
    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        t.Fatal(err)
    }
    defer listener.Close()
    
    go server.Serve(listener)
    
    // 发送测试请求
    conn, err := net.Dial("tcp", listener.Addr().String())
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()
    
    // 发送HTTP请求
    fmt.Fprintf(conn, "GET /hello HTTP/1.1\r\nHost: localhost\r\n\r\n")
    
    // 读取响应
    response := make([]byte, 1024)
    n, err := conn.Read(response)
    if err != nil {
        t.Fatal(err)
    }
    
    responseStr := string(response[:n])
    if !strings.Contains(responseStr, "200 OK") {
        t.Error("expected 200 OK status")
    }
    
    if !strings.Contains(responseStr, "Hello, World!") {
        t.Error("expected response body")
    }
}
```

## 测试工具和辅助函数

### 测试辅助工具
```go
// internal/testing/helpers.go
package testing

// CreateTestServer 创建测试服务器
func CreateTestServer(t *testing.T) (*Server, net.Listener) {
    server := NewServer()
    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        t.Fatal(err)
    }
    return server, listener
}

// SendHTTPRequest 发送HTTP请求
func SendHTTPRequest(t *testing.T, addr, request string) string {
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()
    
    _, err = conn.Write([]byte(request))
    if err != nil {
        t.Fatal(err)
    }
    
    response := make([]byte, 4096)
    n, err := conn.Read(response)
    if err != nil {
        t.Fatal(err)
    }
    
    return string(response[:n])
}
```

### Mock对象
```go
// internal/testing/mock.go
type MockResponseWriter struct {
    headers    map[string]string
    statusCode int
    body       []byte
}

func (m *MockResponseWriter) Header() Header {
    return Header(m.headers)
}

func (m *MockResponseWriter) WriteHeader(code int) {
    m.statusCode = code
}

func (m *MockResponseWriter) Write(data []byte) (int, error) {
    m.body = append(m.body, data...)
    return len(data), nil
}
```

## TDD最佳实践

### 1. 测试驱动设计
- 先写测试，明确接口契约
- 测试即文档，描述预期行为
- 小步迭代，频繁反馈

### 2. 测试组织
- 使用表驱动测试
- 测试用例要覆盖边界条件
- 错误路径同样重要

### 3. 性能考虑
- 基准测试验证性能
- 内存泄漏检测
- 并发安全测试

### 4. 集成测试策略
- 端到端测试验证完整流程
- 使用真实的网络连接
- 测试协议兼容性

## 持续集成配置

### GitHub Actions示例
```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Run Tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Benchmark
      run: go test -bench=. -benchmem ./...
```

## 开发工作流建议

1. **每天开始**：运行所有测试确保基线
2. **功能开发**：红-绿-重构循环
3. **代码提交**：确保所有测试通过
4. **集成测试**：定期运行完整测试套件
5. **性能回归**：监控性能指标变化

采用TDD开发HTTP协议栈的优势：
- **高质量代码**：测试驱动确保代码正确性
- **良好设计**：先考虑接口再实现
- **重构安全**：测试保护重构过程
- **文档化**：测试即规格说明
- **快速反馈**：及时发现问题

这种方法特别适合网络协议实现，因为协议有明确的规范和预期行为，非常适合用测试来描述和验证。