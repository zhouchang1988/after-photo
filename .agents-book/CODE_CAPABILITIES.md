# 代码表现能力清单

> 最后更新：2026-01-01
> 范围：仅基于当前仓库代码、配置、接口定义与 AGENTS.md；不包含人工项目元信息。
> 证据索引：`.agents-book/evidence_index.jsonl`
> 说明：本文是唯一维护的能力清单；JSON 由后续 ingestion 从本文派生，不在 agents-book 中双写。

## 1. 能力总览

| 能力 | 类型 | 代码角色 | 关键证据 | 置信度 |
|---|---|---|---|---|
| 照片类型拆分 | implemented | 扫描目录，按扩展名分类移动文件 | `pkg/step1.go::Step1` | high |
| 重复照片检测与分组 | implemented | pHash + 时间阈值检测相似照片，自动分组 | `pkg/step2.go::Step2` | high |
| 最佳照片选择 | implemented | 清晰度+亮度评分，标记最佳照片 | `pkg/step3.go::Step3` | high |
| 多余 RAW 文件清理 | implemented | 检测无对应 JPG 的 RAW 文件，用户确认后删除 | `pkg/step4.go::Step4` | high |
| TUI 交互界面 | exposed | Bubble Tea TUI 界面，支持步骤选择、执行监控、确认对话 | `main.go::main` | high |
| 日志记录 | implemented | 自动去除 ANSI 码，添加时间戳，写入日志文件 | `main.go::logWriter` | high |
| 用户确认机制 | implemented | 支持 TUI channel 和 CLI 两种确认模式 | `pkg/config.go::RequestConfirm` | high |

## 2. Implemented Capabilities

### 照片类型拆分
- 别名/关键词：文件分类、目录拆分、类型分拣
- Entry Points：`pkg/step1.go::Step1(photoDir)`
- Core Logic：`pkg/step1.go::step1(photoDir)`
- Data Resources：创建 `jpg/`、`raw/`、`video/` 目录
- Side Effects：移动文件到对应目录，删除空的 `video/` 目录
- Evidence：`pkg/step1.go::step1` (line 11), `pkg/config.go::rawExtensions` (line 88), `pkg/config.go::videoExtensions` (line 112)
- Confidence：high
- Limits：仅描述代码表现；不推断业务 owner

### 重复照片检测与分组
- 别名/关键词：去重、相似照片、pHash、感知哈希
- Entry Points：`pkg/step2.go::Step2(photoDir)`
- Core Logic：`pkg/step2.go::step2(photoDir)` → `processDirectory()` → `calculateImageHash()` → `hammingDistance()`
- Algorithm：pHash（32x32 缩放 → DCT → 8x8 低频系数 → 64 位哈希）+ 时间阈值（30 秒）+ 汉明距离（≤17）
- Data Resources：在 `jpg/` 和 `raw/` 下创建分组子目录
- Side Effects：移动文件到分组目录
- Evidence：`pkg/step2.go::calculateImageHash` (line 293), `pkg/step2.go::performDCT` (line 352), `pkg/step2.go::hammingDistance` (line 419), `pkg/config.go::SimilarityThreshold` (line 122)
- Confidence：high
- Limits：仅描述代码表现

### 最佳照片选择
- 别名/关键词：选优、质量评分、清晰度检测
- Entry Points：`pkg/step3.go::Step3(photoDir)`
- Core Logic：`pkg/step3.go::step3(photoDir)` → `selectBestInGroups()` → `calculateImageQualityScore()`
- Algorithm：清晰度（Laplace 算子，权重 0.6）+ 亮度分布（中心/整体比，权重 40）
- Data Resources：重命名最佳文件添加 `-` 后缀
- Side Effects：同步 RAW 标记，删除空的 `jpg/` 和 `raw/` 目录
- Evidence：`pkg/step3.go::calculateImageQualityScore` (line 231), `pkg/step3.go::selectBestInGroups` (line 43), `pkg/step3.go::keepRawByJpgSelection` (line 146)
- Confidence：high
- Limits：仅描述代码表现

### 多余 RAW 文件清理
- 别名/关键词：RAW 清理、存储释放
- Entry Points：`pkg/step4.go::Step4(photoDir)`
- Core Logic：`pkg/step4.go::step4(photoDir)`
- Data Resources：遍历 `raw/` 目录，检查对应 JPG 是否存在
- Side Effects：用户确认后删除 RAW 文件
- Evidence：`pkg/step4.go::step4` (line 11), `pkg/config.go::RequestConfirm` (line 39)
- Confidence：high
- Limits：仅描述代码表现；删除操作需用户确认

### 日志记录
- 别名/关键词：操作日志、执行记录
- Entry Points：`main.go::logWriter`
- Core Logic：`main.go::logWriter::Write()` 自动去除 ANSI 码，添加时间戳
- Data Resources：在工作目录创建 `after_photo_YYYYMMDDHHmmss.txt` 日志文件
- Evidence：`main.go::logWriter` (line 196), `main.go::removeANSICodes` (line 731)
- Confidence：high
- Limits：仅描述代码表现

### 用户确认机制
- 别名/关键词：确认对话、TUI 确认、channel 确认
- Entry Points：`pkg/config.go::RequestConfirm(message)`
- Core Logic：TUI 模式通过 `ConfirmCh` channel 发送确认请求，CLI 模式使用 `defaultConfirm()`
- Evidence：`pkg/config.go::RequestConfirm` (line 39), `pkg/config.go::ConfirmRequest` (line 28), `main.go::waitForConfirmRequest` (line 540)
- Confidence：high
- Limits：仅描述代码表现

## 3. Exposed Capabilities

### TUI 交互界面
- 别名/关键词：终端界面、Bubble Tea、交互式操作
- Entry Points：`main.go::main()` → `tea.NewProgram(initialModel())`
- State Machine：`stateInputDir` → `stateSelectSteps` → `stateRunning` → `stateDone`（+ `stateConfirm`）
- Features：目录输入、步骤选择（复选框）、执行输出滚动、确认对话、完成选项
- Evidence：`main.go::main` (line 715), `main.go::model` (line 220), `main.go::Update` (line 260)
- Confidence：high
- Limits：仅描述代码表现

## 4. Consumed Capabilities

无。本仓库不消费外部事件或消息。

## 5. Delegated / Dependent Capabilities

无。本仓库不调用外部系统。

## 6. Stored Data Capabilities

本仓库不持久化数据到数据库或外部存储。所有操作都是文件系统级别的移动、重命名和删除。

## 7. Evidence Notes And Limits

- 本文不包含 owner / 团队 / 部门 / 联系人 / 权威系统状态
- 这些信息需要由人工元数据或权威系统补充
- 不维护 `CODE_CAPABILITIES.json`，避免 Markdown/JSON 双写不一致；下游系统应从本文解析生成 JSON
- 本项目是纯本地文件操作工具，无外部依赖、无数据库、无消息队列、无 API 调用
