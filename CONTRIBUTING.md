# 贡献指南

感谢您对HTTP协议栈项目的关注！我们欢迎所有形式的贡献。

## 🤝 贡献方式

### 报告Bug
- 使用GitHub Issues报告bug
- 提供详细的重现步骤
- 包含错误日志和环境信息
- 使用适当的标签标记问题

### 提出功能建议
- 通过GitHub Issues提出建议
- 详细描述功能需求和使用场景
- 考虑向后兼容性影响
- 讨论实现方案

### 提交代码
- Fork项目到您的GitHub账户
- 创建功能分支：`git checkout -b feature/amazing-feature`
- 遵循代码规范和测试要求
- 提交PR并填写详细说明

## 📋 开发规范

### 代码风格
- 遵循Go官方代码风格指南
- 使用`gofmt`和`goimports`格式化代码
- 函数和方法必须有文档注释
- 导出的类型、常量、变量必须有文档注释
- 保持代码简洁和可读性

### 命名规范
- 包名：小写，简短，有意义
- 函数名：驼峰命名，首字母大写表示导出
- 变量名：驼峰命名，避免缩写
- 常量名：全大写，下划线分隔
- 接口名：以`-er`结尾（如`Writer`、`Reader`）

### 提交规范

提交消息格式：
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
- `perf`: 性能优化

示例：
```
feat(http): 添加HTTP/2服务器推送功能

实现了服务器推送机制，允许服务器主动向客户端推送资源。
包含完整的测试用例和文档更新。

Closes #123
```

### 测试要求

#### 必须条件
- **测试覆盖率 ≥ 90%**
- 所有公开API必须有单元测试
- 复杂逻辑必须有集成测试
- 性能关键路径必须有基准测试
- 新增功能必须有测试用例

#### 测试分类
1. **单元测试**: 测试单个函数或方法
2. **集成测试**: 测试组件间交互
3. **端到端测试**: 测试完整流程
4. **性能测试**: 基准测试和压力测试
5. **安全测试**: 安全漏洞和边界测试

#### 测试编写指南
```go
func TestFunctionName(t *testing.T) {
    // Arrange - 准备测试数据
    input := "test data"
    expected := "expected result"
    
    // Act - 执行被测试功能
    result := FunctionName(input)
    
    // Assert - 验证结果
    if result != expected {
        t.Errorf("FunctionName(%q) = %q, want %q", input, result, expected)
    }
}
```

#### 基准测试
```go
func BenchmarkFunctionName(b *testing.B) {
    for i := 0; i < b.N; i++ {
        FunctionName("test data")
    }
}
```

## 🔄 开发流程

### 设置开发环境

1. **安装依赖**
```bash
# 安装Go 1.21+
# 安装golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

# 安装goimports
go install golang.org/x/tools/cmd/goimports@latest
```

2. **克隆项目**
```bash
git clone https://github.com/narcilee7/http-stack.git
cd http-stack
go mod download
```

3. **验证环境**
```bash
make test
make lint
```

### 开发步骤

1. **创建功能分支**
```bash
git checkout -b feature/your-feature-name
```

2. **TDD开发流程**
```bash
# 1. 编写失败的测试
# 2. 编写最简单的实现让测试通过
# 3. 重构代码改善设计
# 4. 重复以上步骤
```

3. **提交前检查**
```bash
make test      # 运行所有测试
make lint      # 代码质量检查
make fmt       # 格式化代码
make build     # 构建验证
```

4. **提交代码**
```bash
git add .
git commit -m "feat(component): 简短描述"
git push origin feature/your-feature-name
```

5. **创建Pull Request**
- 填写详细的PR描述
- 链接相关的Issue
- 等待代码审查
- 根据反馈修改

### Pull Request检查清单

- [ ] 代码遵循项目风格指南
- [ ] 添加了适当的测试用例
- [ ] 测试覆盖率达到要求
- [ ] 所有测试都通过
- [ ] 代码通过lint检查
- [ ] 更新了相关文档
- [ ] 提交消息格式正确
- [ ] 没有引入破坏性变更
- [ ] 性能没有明显退化

## 🐛 Bug报告模板

```markdown
**Bug描述**
简要描述遇到的问题。

**重现步骤**
1. 执行 '...'
2. 点击 '....'
3. 滚动到 '....'
4. 看到错误

**期望行为**
描述您期望发生的行为。

**实际行为**
描述实际发生的行为。

**环境信息**
- OS: [e.g. Ubuntu 20.04]
- Go版本: [e.g. 1.21.0]
- 项目版本: [e.g. v0.1.0]

**其他信息**
添加任何其他相关信息、截图、错误日志等。
```

## 💡 功能建议模板

```markdown
**功能描述**
简要描述您建议的功能。

**使用场景**
描述为什么需要这个功能，解决什么问题。

**建议的解决方案**
描述您希望如何实现这个功能。

**替代方案**
描述您考虑过的其他解决方案。

**其他信息**
添加任何其他相关信息或截图。
```

## 📝 代码审查指南

### 审查者指南

- 关注代码逻辑、性能、安全性
- 检查测试覆盖率和质量
- 验证文档和注释
- 确保遵循项目规范
- 给出建设性的反馈

### 被审查者指南

- 详细描述PR的目的和实现
- 主动解释复杂的设计决策
- 积极回应审查意见
- 及时修复发现的问题
- 保持开放和学习的心态

## 🏷️ 版本发布

### 语义化版本控制

我们遵循[语义化版本控制](https://semver.org/)：
- `MAJOR.MINOR.PATCH`
- `MAJOR`: 不兼容的API变更
- `MINOR`: 向后兼容的功能新增
- `PATCH`: 向后兼容的错误修复

### 发布流程

1. 更新版本号和CHANGELOG
2. 创建发布分支
3. 运行完整测试套件
4. 创建Git标签
5. 发布到GitHub Releases
6. 更新文档和示例

## 🤝 社区行为准则

我们致力于创建一个开放、友好、多元化、包容的社区环境。

### 我们的标准

积极行为包括：
- 使用友好和包容的语言
- 尊重不同的观点和经验
- 接受建设性的批评
- 关注对社区最有利的事情
- 对其他社区成员表现出同理心

不当行为包括：
- 使用性化的语言或图像
- 恶意评论或人身攻击
- 公开或私下的骚扰
- 未经许可发布他人私人信息
- 其他在专业环境中不当的行为

### 举报和执行

如果遇到不当行为，请通过GitHub Issues或邮件联系项目维护者。我们会认真对待所有投诉，并采取适当的措施。

## 📞 获取帮助

如果您在贡献过程中遇到任何问题：

1. 查看项目文档和已有Issues
2. 在GitHub Discussions中提问
3. 创建新的Issue寻求帮助
4. 联系项目维护者

感谢您的贡献！🎉