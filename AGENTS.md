# After Photo 开发指南

> 最后更新：2026-01-01
> 位置：`/`

## 1. 系统概述

After Photo 是一个面向连拍摄影师的桌面效率工具，解决"海量重复照片中快速筛选最优解"的问题。

**核心业务边界：**
- 按文件类型拆分目录（JPG/RAW/视频）
- pHash 感知哈希 + 时间阈值检测重复照片
- 基于清晰度和亮度分布自动选优
- 清理无对应 JPG 的多余 RAW 文件

**技术栈：**
- 语言：Go 1.24
- TUI 框架：Bubble Tea（`github.com/charmbracelet/bubbletea`）
- 组件库：Bubbles（`github.com/charmbracelet/bubbles`）
- 样式库：Lip Gloss（`github.com/charmbracelet/lipgloss`）
- 图像处理：标准库 `image` + `image/jpeg`
- 算法：pHash（感知哈希）、DCT（离散余弦变换）、Laplace 算子（清晰度检测）

## 2. 目录导航

| 目录 | 职责 | 关键说明 | AGENTS.md 链接 |
|---|---|---|---|
| `pkg/` | 核心功能包 | 4 个步骤的业务逻辑、配置、图像处理算法 | [./pkg/AGENTS.md](./pkg/AGENTS.md) |
| `bin/` | 编译输出 | macOS/Windows 可执行文件，不纳入版本控制 | 跳过 |
| `test/` | 测试数据 | `input/` 原始测试文件，`output/` 预期结果 | 父级摘要 |

## 3. 核心业务场景索引

- **照片类型拆分**
  - 入口：`pkg.Step1()` → `pkg/step1.go::step1()`
  - 逻辑：扫描目录 → 按扩展名分类 → 移动到 `jpg/`、`raw/`、`video/`

- **重复照片检测与分组**
  - 入口：`pkg.Step2()` → `pkg/step2.go::step2()`
  - 逻辑：按创建时间排序 → 30 秒时间阈值预分组 → pHash 计算 → 汉明距离 ≤ 17 判定相似 → 移动到子目录
  - 算法：`calculateImageHash()` → `performDCT()` → `hammingDistance()`

- **最佳照片选择**
  - 入口：`pkg.Step3()` → `pkg/step3.go::step3()`
  - 逻辑：遍历分组 → `calculateImageQualityScore()` 评分 → 最佳文件添加 `-` 后缀 → 同步 RAW 标记
  - 评分：清晰度（Laplace 算子，权重 0.6）+ 亮度分布（中心/整体比，权重 40）

- **多余 RAW 文件清理**
  - 入口：`pkg.Step4()` → `pkg/step4.go::step4()`
  - 逻辑：遍历 `raw/` → 检查对应 JPG 是否存在 → 用户确认后删除
  - 安全：需用户输入 `y` 确认，支持 TUI channel 确认

- **TUI 交互流程**
  - 入口：`main.go::main()` → `tea.NewProgram(initialModel())`
  - 状态机：`stateInputDir` → `stateSelectSteps` → `stateRunning` → `stateDone`
  - 确认流程：`stateRunning` ↔ `stateConfirm`（step4 删除确认）

## 4. 全局设计约束

- **文件操作**：所有文件移动使用 `os.Rename()`，删除使用 `os.Remove()`
- **输出机制**：全局 `pkg.out` 变量，TUI 模式通过 `channelWriter` 捕获输出到 Bubble Tea
- **确认机制**：`pkg.ConfirmCh` channel 实现 TUI 模式确认，`pkg.confirmFunc` 支持 CLI 模式
- **日志机制**：`logWriter` 自动去除 ANSI 颜色码，添加时间戳前缀
- **禁止事项**：
  - 不要在 `pkg/` 包中直接使用 `fmt.Println()`，必须使用 `fmt.Fprintf(out, ...)`
  - 不要修改 `rawExtensions` 和 `videoExtensions` 映射的键名格式（全大写）

## 5. AGENTS 维护规则

1. **代码变更联动**：任何 Agent 修改了 `pkg/` 中的代码（step1-4.go、config.go），必须同步更新 `pkg/AGENTS.md`
2. **根目录联动**：如果新增/删除了步骤、修改了 TUI 状态机、变更了全局约束，必须更新根目录 `AGENTS.md`
3. **新增目录**：如果新增了业务模块目录（如 `internal/`、`cmd/`），必须为其创建 `AGENTS.md`
4. **清理失效引用**：重命名或删除文件后，必须清理所有 `AGENTS.md` 中的失效引用
5. **跳过文档维护需用户批准**：除非用户明确要求跳过，否则文档维护是任务完成的必要条件
