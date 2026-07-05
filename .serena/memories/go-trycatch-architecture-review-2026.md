# go-trycatch 架构缺陷审查与修复记录

## 项目名称
github.com/shengcnli1982/go-trycatch — Go try-catch-finally 错误处理库

## 发现的缺陷清单

### P0 (3 项，已修复)
1. **Do() panic 路径 catch 不执行 + 返回 nil**
   - 根因: 内层闭包设置 catchCalled=true 后 re-panic，外层 defer 用 `!catchCalled` 判断跳过 catch；Do() 非命名返回值导致 defer 赋值不影响返回
   - 修复: 删除内层闭包，改为命名返回值 `Do() (err error)`，defer 直接 recover + 处理
2. **TryWithResult re-panic**
   - 根因: defer 中 `panic(panicErr)` 导致函数签名 `(T, error)` 失效
   - 修复: 用 `err = panicErr` 替代 `panic(panicErr)`
3. **TryWithResultAndFinally 同上 + finally 顺序错误**
   - 根因: defer 中 finally 在 recover 之前执行
   - 修复: 单个 defer 中先 recover 再 finally

### P1 (2 项，已修复)
4. **Reset() 不完整**: 只清理 try/catch/finally，不清理 ctx/hooks/name（Pool 复用时泄露）
5. **context 取消跳过 finally**: early return 绕过 defer 路径

### 修复技巧
- Go 中 `recover()` 必须在 defer 的顶层函数调用，不能在内层闭包中调用
- 命名返回值是 defer 修改返回值的唯一方式，非命名返回值赋值无效
- 用 IIFE 保护 catch，防止 catch panic 阻止 finally 执行
- catchPanicErr 变量暂存 catch panic，finally 执行后再重新 panic

## 第二轮修复（ERC 收敛轮次 1）

### P0 (1 项，已修复)
1. **Do() nil try 跳过 finally**
   - 根因: `if tc.try == nil { return nil }` 在 defer 注册之前执行，导致 finally/OnFinally 不执行
   - 修复: 将 nil try 检查移到 defer 注册之后（line 116-118），defer 始终注册保证 finally 执行
   - 设计决策: nil try 时不调用 OnTryStart/OnTryEnd（没有 try 可执行），但 finally 必须执行

### P1 (1 项，已修复)
2. **TryCatchBlockWithOptions 死代码**
   - 根因: 结构体已定义但无构造函数/方法/使用
   - 修复: 直接从 options.go 中删除

### 新增测试
- `TestTryCatchBlock_Do_NilTryFinallyExecutes`: nil try 时 finally 仍执行
- `TestTryCatchBlock_Do_NilTryWithOnFinally`: nil try 时 OnFinally 钩子仍执行
- `TestTryCatchBlock_Do_NilTryReturnsNil`: nil try 时返回 nil

### 最终状态
- 39 个测试全部通过（含 -race）
- 代码覆盖率 96.5%
- go vet + go build 零错误
- ERC 收敛：完成度100%, P0=0, P1=0, 质量88分

## 测试反模式修复
- 断言放在闭包内（panic 场景下闭包不被执行 → 假 PASS）
- 应移到闭包外验证，或用闭包外变量追踪
- goroutine 内不用 assert/t.FailNow（会导致 goroutine panic），改用 atomic 计数器

## 设计一致性决策
- TryWithResultAndFinally 的 finally panic 应向上传播（与 Do() 的 FinallyPanic 行为一致），非设计缺陷
