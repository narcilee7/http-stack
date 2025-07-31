// Package main 提供全局测试配置
package main

import (
	"os"
	"testing"
)

// TestMain 测试主函数，用于全局测试设置和清理
func TestMain(m *testing.M) {
	// 全局测试设置
	setup()

	// 运行测试
	code := m.Run()

	// 全局测试清理
	teardown()

	os.Exit(code)
}

// setup 全局测试设置
func setup() {
	// 设置测试环境变量
	os.Setenv("HTTP_STACK_TEST_MODE", "true")
	os.Setenv("HTTP_STACK_LOG_LEVEL", "debug")
}

// teardown 全局测试清理
func teardown() {
	// 清理测试环境变量
	os.Unsetenv("HTTP_STACK_TEST_MODE")
	os.Unsetenv("HTTP_STACK_LOG_LEVEL")
}
