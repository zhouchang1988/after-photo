# GUI 模块开发指南

> 最后更新：2026-07-29（GUI 整体重写）
> 位置：`gui/`

## 1. 模块概述

gui 模块是 After Photo 的桌面 GUI 版本，基于 Wails v2 框架构建。它将 `pkg/` 包的业务逻辑封装为桌面应用，提供图形界面操作、实时日志、进度条、目录预览统计、任务取消与删除确认等能力。

**核心职责：**
- 封装 `pkg.Step1-4()` 为 GUI 可调用的方法
- 目录校验与文件统计预览（`ScanDirectory`）
- 实时日志流（保留 ANSI 颜色，由前端渲染）
- 步骤内进度推送（`pkg.SetProgressFunc` → `progress` 事件）
- step4 删除确认对话框、任务取消
- 日志文件写入（去除 ANSI 码，添加时间戳）

**技术栈：**
- 框架：Wails v2（`github.com/wailsapp/wails/v2`）
- 后端：Go（依赖 `after_photo/pkg`）
- 前端：原生 HTML/CSS/JavaScript（无框架、无构建步骤）
- 构建：Wails CLI（`wails build`），`wails.json` 中 install/build 命令为空

## 2. 核心代码结构

| 文件/目录 | 职责 | 关键类型/函数 |
|---|---|---|
| `app.go` | 后端全部逻辑 | `App`, `ScanDirectory()`, `StartProcessing()`, `CancelProcessing()`, `ConfirmStep4()`, `logWriter`, `stripANSI()` |
| `main.go` | Wails 应用入口 | `main()`, `wails.Run()`（窗口 1180×800，深色背景） |
| `frontend/index.html` | 页面结构 | 左侧配置面板 + 右侧日志视图 + 确认弹窗 |
| `frontend/src/main.js` | 前端逻辑 | 状态机、ANSI 渲染、Wails 事件订阅 |
| `frontend/src/style.css` | 样式 | CSS 变量实现深/浅双主题（深色默认） |
| `frontend/dist/` | 构建输出 | **手动同步**自 `frontend/`（无构建命令），嵌入 Go 二进制 |
| `frontend/wailsjs/` | Wails 绑定 | 由 `wails build` 自动生成 |

## 3. 后端 API（window.go.main.App.*）

| 方法 | 说明 |
|---|---|
| `GetDefaultDir()` | 可执行文件所在目录 |
| `PickDirectory()` | 系统目录选择框 |
| `ScanDirectory(dir)` | 校验目录并统计 JPG/RAW/视频/其他数量与总大小，返回 `DirStats` |
| `StartProcessing(dir, steps [4]bool)` | 后台 goroutine 顺序执行选中步骤，立即返回 |
| `CancelProcessing()` | 请求取消（当前步骤结束后停止，pkg 步骤不可中断） |
| `ConfirmStep4(confirmed bool)` | 回应 step4 确认请求 |
| `Quit()` | 退出应用 |

> **注意**：Wails 生成的 JS 绑定完整保留方法名（含末尾数字），前端调用 `window.go.main.App.ConfirmStep4(...)`，与 Go 方法名一致。

## 4. Wails 事件（后端 → 前端）

| 事件 | 载荷 | 说明 |
|---|---|---|
| `log` | `string` | pkg 实时输出（含 ANSI 颜色码，前端 `ansiToHtml()` 渲染） |
| `progress` | `{current, total, message}` | 步骤内进度（来自 `pkg.SetProgressFunc`），total 为 0 时前端显示不确定进度 |
| `step` | `{index, total, name}` | 某步骤开始执行 |
| `confirm` | `string` | step4 删除确认请求，前端弹窗后调用 `ConfirmStep` 回应 |
| `done` | `{duration, cancelled}` | 全部步骤结束（含取消与 panic 恢复） |

## 5. 核心业务流程

```
选择目录 → ScanDirectory（校验 + 统计，显示在左侧）
  → 勾选步骤 → 开始处理
  → StartProcessing
      → 创建日志文件 after_photo_YYYYMMDDHHmmss.txt（工作目录下；每轮运行前关闭上一轮句柄）
      → pkg.SetOutput(App)（App 实现 io.Writer，Write 转发为 log 事件 + 写日志文件）
      → pkg.SetProgressFunc / pkg.SetConfirmFunc
      → goroutine 逐步执行，每步前发 step 事件
  → step4：confirm 事件 → 前端弹窗 → ConfirmStep(true/false)
  → done 事件：恢复 UI、停止计时
```

### 确认机制（重写后）

```
pkg.RequestConfirm → SetConfirmFunc 注入的函数
  → 将 result chan 存入 App.confirmRes（mutex 保护）
  → 发 confirm 事件并阻塞等待
  → 前端弹窗 → ConfirmStep(ok) → 从 confirmRes 取出 chan 发送结果
```

相比旧实现：不再用 `confirmCh + 5 秒超时轮询`，改为「存请求、直接应答」，无超时竞态；应用退出时通过 `ctx.Done()` 解除阻塞。

## 6. 前端设计要点

- **布局**：左侧 320px 配置面板（品牌、目录+统计、步骤卡片、运行按钮），右侧日志视图（工具栏 + 控制台 + 状态栏）
- **主题**：`[data-theme]` CSS 变量，深/浅双主题，localStorage 持久化；旧版 33 套主题已移除
- **ANSI 渲染**：`ansiToHtml()` 处理颜色码 30-37 / 加粗 / 重置，支持转义序列跨 chunk 留存（`ansiPending`）
- **控制台**：DOM 节点超过 4000 自动裁剪头部，防止大目录日志拖垮页面
- **步骤卡片**：on（选中）/ running（转圈）/ finished（对勾）状态，由 `step` 与 `done` 事件驱动
- **日志句柄**：`StartProcessing` 每次运行前先关闭上一轮的日志文件句柄，避免跨轮泄漏
- **退出保护**：`shutdown` 中 `wg.Wait()` 限时 2 秒——pkg 步骤不可中断，长时间步骤（如大批量哈希）不能卡住关窗流程
- **dist 同步**：修改 `frontend/index.html` 或 `frontend/src/*` 后，必须手动复制到 `frontend/dist/`：
  `cp frontend/index.html frontend/dist/ && cp frontend/src/*.js frontend/src/*.css frontend/dist/src/`

## 7. 测试入口

- **编译**：`wails build`（输出 `gui/build/bin/After Photo.app`）或根目录 `build.sh --gui`
- **冒烟**：直接运行 `gui/build/bin/After Photo.app/Contents/MacOS/after-photo-gui`，确认进程存活
- **开发**：`wails dev` 热重载

## 8. 维护与风险说明

| 风险 | 概率 | 影响 | 缓解 |
|---|---|---|---|
| 改前端后忘记同步 `dist/` | 中 | 界面不更新 | 见 6 节同步命令；构建前检查 |
| 前端方法名与 Go 不一致 | 低 | 调用 undefined | 以 `frontend/wailsjs/go/main/App.js` 生成结果为准核对命名 |
| pkg 步骤不可中断 | 中 | 取消/退出延迟 | UI 提示「等待当前步骤结束」；shutdown 限时 2 秒等待后强制退出 |
| 扫描扩展名与 pkg 不一致 | 低 | 统计偏差 | `app.go` 中 `jpegExts/rawExts/videoExts` 需与 pkg 同步维护 |
| 并发执行 | 低 | 数据竞争 | `App.mu` + `processing` 标志保护 |

**下游影响**：修改 `app.go` 导出方法会影响前端调用与 `frontend/wailsjs/`（`wails build` 自动重新生成）。

## 9. 与 TUI 版本的关系

- **共享核心逻辑**：GUI 和 TUI 都调用 `pkg.Step1-4()`
- **不同的交互层**：GUI 使用 Wails + HTML/CSS/JS，TUI 使用 Bubble Tea
- **不同的输出机制**：GUI 通过 `runtime.EventsEmit()` 推送，TUI 通过 `channelWriter` 捕获
- **不同的确认机制**：GUI 使用 Wails 事件 + channel，TUI 使用 Bubble Tea 命令
