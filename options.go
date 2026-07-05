package gotrycatch

import (
	"context"
)

// Option 定义 TryCatchBlock 的配置选项
type Option func(*TryCatchBlock)

// WithContext 添加 context 支持，用于取消和超时控制
func WithContext(ctx context.Context) Option {
	return func(tc *TryCatchBlock) {
		tc.ctx = ctx
	}
}

// Hooks 定义用于监控 TryCatchBlock 执行的回调
type Hooks struct {
	OnTryStart func()      // 在 try 执行前调用
	OnTryEnd   func(error) // 在 try 执行后调用，传入错误结果
	OnCatch    func(error) // 在 catch 执行时调用
	OnFinally  func()      // 在 finally 执行时调用
}

// WithHooks 添加监控执行的钩子
func WithHooks(hooks Hooks) Option {
	return func(tc *TryCatchBlock) {
		tc.hooks = hooks
	}
}

// WithName 为 try-catch 块添加名称标识符
func WithName(name string) Option {
	return func(tc *TryCatchBlock) {
		tc.name = name
	}
}

// Context 返回与 TryCatchBlock 关联的 context
func (tc *TryCatchBlock) Context() context.Context {
	return tc.ctx
}

// Name 返回与 TryCatchBlock 关联的名称
func (tc *TryCatchBlock) Name() string {
	return tc.name
}

// Hooks 返回与 TryCatchBlock 关联的钩子
func (tc *TryCatchBlock) Hooks() Hooks {
	return tc.hooks
}

// ApplyOptions 将提供的选项应用到 TryCatchBlock
func (tc *TryCatchBlock) ApplyOptions(opts ...Option) *TryCatchBlock {
	for _, opt := range opts {
		opt(tc)
	}
	return tc
}
