# pkg 核心功能包开发指南

> 最后更新：2026-01-01
> 位置：`pkg/`

## 1. 模块概述

pkg 包实现了 After Photo 的全部核心业务逻辑：照片分类、重复检测、最佳选择、RAW 清理。这是一个纯业务逻辑包，不包含 UI 代码，通过全局 `out` 变量输出日志，支持 TUI 和 CLI 两种模式。

## 2. 核心代码结构

| 文件 | 职责 | 关键函数/类型 |
|---|---|---|
| `config.go` | 全局配置、输出控制、确认机制 | `SetOutput()`, `RequestConfirm()`, `ConfirmCh`, `rawExtensions`, `videoExtensions`, `SimilarityThreshold` |
| `step1.go` | 按文件类型拆分目录 | `step1()`, `Step1()` |
| `step2.go` | 重复照片检测与分组 | `step2()`, `Step2()`, `processDirectory()`, `processRawByJpgStructure()`, `calculateImageHash()`, `performDCT()`, `hammingDistance()` |
| `step3.go` | 最佳照片选择与标记 | `step3()`, `Step3()`, `selectBestInGroups()`, `keepRawByJpgSelection()`, `calculateImageQualityScore()` |
| `step4.go` | 多余 RAW 文件清理 | `step4()`, `Step4()` |
| `step_test.go` | 单元测试 | 测试各步骤函数 |

## 3. 核心业务流程

### 步骤 1：文件类型拆分
```
step1(photoDir)
  → os.MkdirAll("jpg/", "raw/", "video/")
  → os.ReadDir(photoDir)
  → 按扩展名分类：JPG/JPEG → jpg/ | RAW扩展名 → raw/ | 视频扩展名 → video/
  → os.Rename(filePath, destPath)
  → 删除空的 video/ 目录
```

### 步骤 2：重复照片检测
```
step2(photoDir)
  → processDirectory(jpgDir, "JPG")
      → filepath.Walk 收集 JPG 文件
      → 按 birthTime 排序
      → 30 秒时间阈值预分组
      → calculateImageHash() 计算 pHash
      → hammingDistance() ≤ 17 判定相似
      → moveGroup() 移动到子目录
  → processRawByJpgStructure(jpgDir, rawDir)
      → 按 JPG 分组结构同步移动 RAW 文件
```

### 步骤 3：最佳照片选择
```
step3(photoDir)
  → selectBestInGroups(jpgDir, "JPG")
      → filepath.Walk 遍历分组目录
      → calculateImageQualityScore() 评分
      → 最佳文件重命名添加 "-" 后缀
  → removeEmptyDir(jpgDir) 删除空目录
  → keepRawByJpgSelection(jpgDir, rawDir)
      → 同步 RAW 文件的 "-" 标记
  → removeEmptyDir(rawDir) 删除空目录
```

### 步骤 4：多余 RAW 清理
```
step4(photoDir)
  → filepath.Walk(rawDir) 收集所有 RAW 文件
  → 检查对应 JPG 是否存在（.JPG 或 .jpg）
  → 收集 filesToDelete 列表
  → RequestConfirm() 用户确认
  → os.Remove() 执行删除
```

## 4. 关键资源与副作用

- **文件系统操作**：
  - 创建目录：`jpg/`、`raw/`、`video/` 及分组子目录
  - 移动文件：`os.Rename()` 重命名/移动
  - 删除文件：`os.Remove()` 删除多余 RAW（需用户确认）
  - 删除目录：`os.RemoveAll()` 删除空目录

- **全局状态**：
  - `out io.Writer`：全局输出 writer，TUI 模式注入 channelWriter
  - `confirmFunc`：确认函数，TUI 模式通过 `ConfirmCh` channel 替换
  - `ConfirmCh chan *ConfirmRequest`：TUI 确认请求 channel

## 5. 常见修改场景与切入点

- **新增文件类型支持**：修改 `config.go` 中的 `rawExtensions` 或 `videoExtensions` 映射
- **调整相似度阈值**：修改 `config.go` 中的 `SimilarityThreshold`（当前值 17）
- **修改评分算法**：修改 `step3.go` 中的 `calculateImageQualityScore()`，调整清晰度/亮度权重
- **调整时间阈值**：修改 `step2.go` 中的 `const timeThreshold = 30`
- **修改分组命名规则**：修改 `step2.go` 中的 `moveGroup()` 函数
- **修改确认行为**：修改 `config.go` 中的 `RequestConfirm()` 和 `defaultConfirm()`

## 6. 测试入口

- 单元测试：`pkg/step_test.go`（测试各步骤函数）
- 集成测试：`main_test.go`（TUI 模型初始化、channelWriter、logWriter、ANSI 码去除）
- 端到端测试：`build.sh --test`（使用 `test/input` 和 `test/output` 比较结果）

## 7. 维护与风险说明

| 风险 | 概率 | 影响 | 缓解 |
|---|---|---|---|
| 文件移动失败 | 低 | 用户数据丢失 | 使用 `os.Rename()` 原子操作，失败时输出错误信息 |
| 误删 RAW 文件 | 中 | 用户数据丢失 | step4 需用户确认，步骤默认不选中 |
| pHash 误判 | 低 | 分组错误 | 阈值 17 经过测试验证，30 秒时间阈值减少误判 |
| 空目录残留 | 低 | 目录混乱 | step3 自动清理空的 jpg/ 和 raw/ 目录 |

**下游影响**：修改 `pkg/` 中的任何函数都会直接影响 TUI 界面的执行结果。`Step1-4()` 是 `main.go` 的直接调用入口。
