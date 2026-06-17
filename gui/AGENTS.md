# GUI 模块开发指南

> 最后更新：2026-06-11
> 位置：`gui/`

## 1. 模块概述

gui 模块是 After Photo 的桌面 GUI 版本，基于 Wails v2 框架构建。它将 `pkg/` 包的业务逻辑封装为桌面应用，提供图形界面操作、实时输出显示、主题切换等功能。

**核心职责：**
- 封装 `pkg.Step1-4()` 为 GUI 可调用的方法
- 提供目录选择、步骤配置、确认对话框等交互
- 实时输出日志到前端界面
- 管理日志文件写入（去除 ANSI 码，添加时间戳）

**技术栈：**
- 框架：Wails v2（`github.com/wailsapp/wails/v2`）
- 后端：Go（依赖 `after_photo/pkg`）
- 前端：原生 HTML/CSS/JavaScript（无框架）
- 构建：Wails CLI（`wails build`）

## 2. 核心代码结构

| 文件/目录 | 职责 | 关键类型/函数 |
|---|---|---|
| `app.go` | 后端核心逻辑 | `App`, `NewApp()`, `StartProcessing()`, `ConfirmStep4()`, `GetOutput()`, `logWriter`, `removeANSICodes()` |
| `main.go` | Wails 应用入口 | `main()`, `wails.Run()` |
| `go.mod` | 模块定义 | 依赖 `after_photo/pkg`, `wails/v2` |
| `wails.json` | Wails 配置 | 应用名称、前端目录、输出文件名 |
| `frontend/` | 前端资源 | HTML、CSS、JavaScript |
| `frontend/index.html` | HTML 入口 | 页面结构 |
| `frontend/src/main.js` | JavaScript 入口 | 前端逻辑、事件监听 |
| `frontend/src/style.css` | 样式 | 33 套主题（20 深色 + 13 浅色） |
| `frontend/dist/` | 构建输出 | 嵌入到 Go 二进制文件 |
| `frontend/wailsjs/` | Wails 绑定 | 自动生成的 Go-JS 桥接代码 |

## 3. 核心业务流程

### GUI 处理流程
```
用户操作
  → 选择目录 (PickDirectory) 或输入路径
  → 验证目录 (ValidateDirectory)
  → 选择步骤 (steps [4]bool)
  → 点击开始
  → StartProcessing(dir, steps)
      → 创建日志文件 (after_photo_YYYYMMDDHHmmss.txt)
      → 设置输出 writer (io.MultiWriter)
      → 设置确认函数 (SetConfirmFunc)
      → 启动 goroutine 执行 pkg.Step1-4
      → 实时输出到前端 (runtime.EventsEmit)
  → step4 需要确认
      → 前端收到 confirm-request 事件
      → 用户点击确认/取消
      → ConfirmStep4(confirmed bool)
  → 执行完成
      → 前端收到 processing-complete 事件
```

### 确认机制
```
pkg.RequestConfirm(message)
  → 调用 SetConfirmFunc 注入的函数
  → 发送 confirmRequest 到 confirmCh
  → 触发前端 confirm-request 事件
  → 前端显示确认对话框
  → 用户点击确认/取消
  → 调用 ConfirmStep4(confirmed)
  → 从 confirmCh 读取请求，发送结果
  → pkg 继续执行
```

## 4. 关键资源与副作用

- **文件系统操作**：
  - 创建日志文件：`after_photo_YYYYMMDDHHmmss.txt`（在用户选择的目录中）
  - 读取/移动/删除照片文件（通过 `pkg` 包）

- **全局状态**：
  - `App.processing bool`：处理状态标志，防止并发执行
  - `App.confirmCh chan *confirmRequest`：确认请求 channel
  - `App.outputBuf *bytes.Buffer`：输出缓冲区
  - `App.logFile *os.File`：日志文件句柄

- **Wails 事件**：
  - `output`：实时输出文本
  - `confirm-request`：确认请求（step4）
  - `processing-complete`：处理完成

## 5. 常见修改场景与切入点

- **修改窗口配置**：修改 `main.go` 中的 `wails.Run()` 参数（标题、尺寸、最小尺寸）
- **添加新的前端事件**：在 `app.go` 中使用 `runtime.EventsEmit()` 发送，在 `frontend/src/main.js` 中监听
- **修改确认对话框行为**：修改 `ConfirmStep4()` 函数或前端确认逻辑
- **添加新的 GUI 方法**：在 `app.go` 中添加方法，在 `main.go` 的 `Bind` 中注册
- **修改主题**：修改 `frontend/src/style.css` 中的 CSS 变量
- **修改日志格式**：修改 `logWriter.Write()` 函数
- **添加新的配置项**：在 `App` 结构体中添加字段，暴露给前端

## 6. 前端结构

### 嵌入机制
```go
//go:embed all:frontend/dist
var assets embed.FS
```
- `frontend/dist/` 目录在构建时嵌入到 Go 二进制文件
- 开发时使用 `wails dev` 实时热重载

### Wails 绑定
- `frontend/wailsjs/` 目录自动生成
- 包含 Go 方法的 JavaScript 绑定
- 前端可直接调用 `window.go.main.App.MethodName()`

## 7. 测试入口

- **手动测试**：`wails dev` 启动开发模式
- **构建测试**：`wails build` 编译为可执行文件
- **集成测试**：使用 `build.sh --gui` 编译 GUI 版本

## 8. 维护与风险说明

| 风险 | 概率 | 影响 | 缓解 |
|---|---|---|---|
| Wails 版本升级破坏 API | 低 | 编译失败 | 锁定 `wails/v2 v2.12.0`，升级前测试 |
| 前端嵌入失败 | 低 | 应用无法启动 | 确保 `frontend/dist/` 存在且非空 |
| 确认请求超时 | 低 | 步骤4 无法执行 | 5 秒超时后返回错误，可重试 |
| 并发处理 | 中 | 数据竞争 | `processingMu` 互斥锁保护 |

**下游影响**：修改 `app.go` 中的方法签名会影响前端 JavaScript 绑定，需要同步更新 `frontend/wailsjs/`（自动生成）。

## 9. 与 TUI 版本的关系

- **共享核心逻辑**：GUI 和 TUI 都调用 `pkg.Step1-4()`
- **不同的交互层**：GUI 使用 Wails + HTML/CSS/JS，TUI 使用 Bubble Tea
- **不同的输出机制**：GUI 通过 `runtime.EventsEmit()` 推送，TUI 通过 `channelWriter` 捕获
- **不同的确认机制**：GUI 使用 Wails 事件 + channel，TUI 使用 Bubble Tea 命令
