#!/bin/bash

set -e

RUN_TEST=false
BUILD_GUI=false
BUILD_TUI=true

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --test|-t)
            RUN_TEST=true
            shift
            ;;
        --gui|-g)
            BUILD_GUI=true
            BUILD_TUI=false
            shift
            ;;
        --all|-a)
            BUILD_GUI=true
            BUILD_TUI=true
            shift
            ;;
        --help|-h)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --test, -t    编译后运行测试"
            echo "  --gui, -g     仅编译 GUI 版本"
            echo "  --all, -a     编译 TUI + GUI 版本"
            echo "  --help, -h    显示帮助"
            echo ""
            echo "示例:"
            echo "  $0            # 编译 TUI 版本"
            echo "  $0 --gui      # 编译 GUI 版本"
            echo "  $0 --all      # 编译所有版本"
            echo "  $0 --test     # 编译并运行测试"
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            echo "使用 --help 查看帮助"
            exit 1
            ;;
    esac
done

# 创建 bin 目录
mkdir -p bin

# 编译 TUI 版本
if [ "$BUILD_TUI" = true ]; then
    echo "=========================================="
    echo "编译 TUI 版本..."
    echo "=========================================="

    # macOS
    echo ""
    echo "编译 macOS 版本..."
    GOOS=darwin GOARCH=amd64 go build -o bin/after-photo-mac
    if [ $? -eq 0 ]; then
        echo "✓ macOS 版本编译成功: bin/after-photo-mac"
    else
        echo "✗ macOS 版本编译失败"
        exit 1
    fi

    echo ""
    echo "TUI 编译完成！"
    ls -lh bin/
fi

# 编译 GUI 版本
if [ "$BUILD_GUI" = true ]; then
    echo ""
    echo "=========================================="
    echo "编译 GUI 版本..."
    echo "=========================================="

    # 检查 wails 是否安装
    if ! command -v wails &> /dev/null; then
        if [ -f "$(go env GOPATH)/bin/wails" ]; then
            export PATH=$PATH:$(go env GOPATH)/bin
        else
            echo "✗ 未找到 wails CLI"
            echo "  请先安装: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
            exit 1
        fi
    fi

    cd gui
    wails build
    if [ $? -eq 0 ]; then
        echo ""
        echo "✓ GUI 版本编译成功"
        echo "  macOS: gui/build/bin/After Photo.app"
    else
        echo "✗ GUI 版本编译失败"
        exit 1
    fi
    cd ..
fi

# 运行测试
if [ "$RUN_TEST" = true ]; then
    echo ""
    echo "=========================================="
    echo "运行测试..."
    echo "=========================================="

    # 备份 input 目录
    TEST_BACKUP_DIR="test/input_backup_$(date +%s)"
    echo "备份 input 目录到 $TEST_BACKUP_DIR..."
    cp -r test/input "$TEST_BACKUP_DIR"

    # 运行程序处理 input
    echo ""
    echo "使用编译的程序处理 input 目录..."
    (echo "test/input"; echo ""; echo "3") | ./bin/after-photo-mac > /dev/null

    # 比较文件结构
    echo ""
    echo "比较文件结构..."

    OUTPUT_FILES=$(find test/output -type f -not -name '.DS_Store' -not -name '*.txt' | sort)
    INPUT_FILES=$(find test/input -type f -not -name '.DS_Store' -not -name '*.txt' | sort)

    OUTPUT_NAMES=$(echo "$OUTPUT_FILES" | sed 's|test/output/||' | sort)
    INPUT_NAMES=$(echo "$INPUT_FILES" | sed 's|test/input/||' | sort)

    if [ "$OUTPUT_NAMES" = "$INPUT_NAMES" ]; then
        FILE_COUNT=$(echo "$OUTPUT_NAMES" | grep -v '^$' | wc -l | tr -d ' ')
        echo "✓ 测试通过！目录结构完全一致 ($FILE_COUNT 个文件)"
        TEST_RESULT=0
    else
        echo "✗ 测试失败！目录结构不一致"
        echo "期望的文件:"
        echo "$INPUT_NAMES"
        echo ""
        echo "实际的文件:"
        echo "$OUTPUT_NAMES"
        TEST_RESULT=1
    fi

    # 恢复 input 目录
    echo ""
    echo "恢复 input 目录..."
    rm -rf test/input
    cp -r "$TEST_BACKUP_DIR" test/input
    rm -rf "$TEST_BACKUP_DIR"

    echo ""
    echo "=========================================="
    echo "测试完成"
    echo "=========================================="

    exit $TEST_RESULT
else
    echo ""
    echo "=========================================="
    echo "编译完成"
    echo ""
    echo "运行方式:"
    echo "  TUI: ./bin/after-photo-mac"
    echo "  GUI: open gui/build/bin/After\\ Photo.app"
    echo ""
    echo "提示:"
    echo "  --test    运行测试"
    echo "  --gui     仅编译 GUI"
    echo "  --all     编译所有版本"
    echo "=========================================="
    exit 0
fi
