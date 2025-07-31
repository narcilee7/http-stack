# Go HTTP协议栈

[![CI](https://github.com/narcilee7/http-stack/actions/workflows/ci.yml/badge.svg)](https://github.com/narcilee7/http-stack/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/narcilee7/http-stack)](https://goreportcard.com/report/github.com/narcilee7/http-stack)
[![codecov](https://codecov.io/gh/narcilee7/http-stack/branch/main/graph/badge.svg)](https://codecov.io/gh/narcilee7/http-stack)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

一个从零开始实现的高性能HTTP协议栈，支持HTTP/1.1和HTTP/2，用于深入学习网络协议和Go高级编程。

## 🎯 项目目标

- **深入理解HTTP协议**：从TCP层到应用层的完整实现
- **高性能设计**：支持C10K问题，QPS ≥ 100K
- **协议完整性**：完全兼容HTTP/1.1和HTTP/2标准
- **安全特性**：内置TLS 1.2/1.3支持和安全防护
- **生产就绪**：工业级代码质量和可靠性

## 🚀 特性

### 核心功能
- [x] HTTP/1.1协议完整支持
- [ ] HTTP/2协议支持（开发中）
- [x] 高效的TCP连接管理
- [x] 连接池和Keep-Alive
- [ ] TLS/HTTPS支持（开发中）
- [ ] 中间件系统（开发中）

### 性能特性
- [x] 零拷贝缓冲区管理
- [x] 内存池化优化
- [ ] 请求管线化（开发中）
- [ ] 流量控制（开发中）
- [ ] 压缩支持（开发中）

### 安全特性
- [ ] TLS 1.2/1.3支持（开发中）
- [ ] 安全头部自动设置（开发中）
- [ ] 请求验证和限制（开发中）
- [ ] CORS支持（开发中）

## 📦 安装

```bash
go get github.com/narcilee7/http-stack
```

## 🔧 快速开始

### HTTP服务器示例

```go
package main

import (
    "github.com/narcilee7/http-stack/pkg/http/server"
)

func main() {
    // 创建服务器实例
    srv := server.New(&server.Config{
        Addr: ":8080",
    })
    
    // 注册路由
    srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, HTTP Stack!"))
    })
    
    // 启动服务器
    srv.ListenAndServe()
}
```

### HTTP客户端示例

```go
package main

import (
    "github.com/narcilee7/http-stack/pkg/http/client"
)

func main() {
    // 创建客户端
    c := client.New(&client.Config{
        Timeout: 30 * time.Second,
    })
    
    // 发送请求
    resp, err := c.Get("http://localhost:8080/")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    
    // 处理响应
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}
```

## 🏗️ 架构设计

```
├── cmd/                    # 应用程序入口
│   ├── server/            # HTTP服务器
│   └── client/            # HTTP客户端
├── pkg/                   # 核心包
│   ├── http/              # HTTP协议实现
│   │   ├── message/       # HTTP消息处理
│   │   ├── protocol/      # 协议层实现
│   │   ├── server/        # 服务器实现
│   │   └── client/        # 客户端实现
│   ├── tcp/               # TCP连接管理
│   ├── tls/               # TLS/SSL支持
│   ├── cache/             # 缓存系统
│   ├── compression/       # 压缩算法
│   └── utils/             # 工具函数
├── internal/              # 内部包
│   ├── config/            # 配置管理
│   ├── metrics/           # 监控指标
│   └── testing/           # 测试工具
└── test/                  # 测试套件
    ├── unit/              # 单元测试
    ├── integration/       # 集成测试
    └── performance/       # 性能测试
```

## 🛠️ 开发指南

### 开发环境要求

- Go 1.21+
- Git
- Make
- golangci-lint

### 克隆项目

```bash
git clone https://github.com/narcilee7/http-stack.git
cd http-stack
```

### 构建项目

```bash
# 安装依赖
go mod download

# 构建服务器
go build -o bin/server ./cmd/server

# 构建客户端  
go build -o bin/client ./cmd/client
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告
go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# 查看覆盖率报告
go tool cover -html=coverage.txt
```

### 代码检查

```bash
# 运行linter
golangci-lint run

# 格式化代码
go fmt ./...

# 自动修复导入
goimports -w .
```

### 性能测试

```bash
# 运行基准测试
go test -bench=. -benchmem ./...

# 生成性能分析报告
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
```

## 📋 开发规范

### 代码风格

- 遵循Go官方代码风格指南
- 使用`gofmt`和`goimports`格式化代码
- 函数和方法必须有文档注释
- 导出的类型、常量、变量必须有文档注释

### 提交规范

```
<类型>(<范围>): <描述>

[可选的正文]

[可选的脚注]
```

类型：
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 其他杂项

### 测试要求

- **测试覆盖率 ≥ 90%**
- 所有公开API必须有单元测试
- 复杂逻辑必须有集成测试
- 性能关键路径必须有基准测试
- 使用TDD方式开发新功能

### Pull Request流程

1. Fork项目并创建功能分支
2. 编写代码和测试
3. 确保所有测试通过
4. 提交PR并填写详细说明
5. 等待代码审查
6. 根据反馈修改代码
7. 合并到主分支

## 📊 性能指标

### 目标性能

- **QPS**: ≥ 100K (单核)
- **延迟**: P99 ≤ 10ms
- **内存使用**: ≤ 100MB (10K连接)
- **CPU使用**: ≤ 80% (峰值负载)

### 当前性能

> 性能测试正在进行中，数据将在第8周更新

## 📈 开发进度

### 第1阶段：项目初始化和基础工具 (第1-2周)
- [x] 项目结构初始化
- [x] CI/CD配置
- [ ] 测试框架搭建
- [ ] 基础工具开发

### 第2阶段：HTTP消息解析 (第3-4周)
- [ ] HTTP头部处理
- [ ] URL解析器
- [ ] Cookie处理
- [ ] 请求/响应消息

> 详细进度请查看 [RoadMap](docs/RoadMap.md)

## 🤝 贡献指南

我们欢迎任何形式的贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细信息。

### 贡献方式

- 提交bug报告
- 提出功能建议
- 改进文档
- 提交代码补丁
- 分享使用经验

## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。

## 📞 联系方式

- **GitHub Issues**: [项目Issues](https://github.com/narcilee7/http-stack/issues)
- **讨论区**: [GitHub Discussions](https://github.com/narcilee7/http-stack/discussions)

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

## 📚 相关资源

- [HTTP/1.1 RFC 7230-7235](https://tools.ietf.org/html/rfc7230)
- [HTTP/2 RFC 7540](https://tools.ietf.org/html/rfc7540)
- [TLS 1.3 RFC 8446](https://tools.ietf.org/html/rfc8446)
- [Go网络编程](https://golang.org/pkg/net/)
- [项目架构设计](docs/File_Design.md)
- [TDD开发方法](docs/TDD.md)

---

⭐ 如果这个项目对你有帮助，请给我们一个星标！

