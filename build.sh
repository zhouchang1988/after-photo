#!/bin/bash

# 测试标志，默认为 false
RUN_TEST=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --test|-t)
            RUN_TEST=true
            shift
            ;;
        *)
            echo "未知参数: $1"
            echo "用法: $0 [--test|-t]"
            exit 1
            ;;
    esac
done

# 创建bin目录
mkdir -p bin

echo "开始编译跨平台版本..."

# 编译 macOS 版本
echo ""
echo "编译 macOS 版本..."
GOOS=darwin GOARCH=amd64 go build -o bin/after-photo-mac
if [ $? -eq 0 ]; then
    echo "✓ macOS 版本编译成功: bin/after-photo-mac"
else
    echo "✗ macOS 版本编译失败"
    exit 1
fi

# 编译 Windows 版本
echo ""
echo "编译 Windows 版本..."
GOOS=windows GOARCH=amd64 go build -o bin/after-photo.exe
if [ $? -eq 0 ]; then
    echo "✓ Windows 版本编译成功: bin/after-photo.exe"
else
    echo "✗ Windows 版本编译失败"
    exit 1
fi

echo ""
echo "所有平台编译完成！"
echo "生成文件:"
ls -lh bin/

# 如果启用了测试，则运行测试
if [ "$RUN_TEST" = true ]; then
    echo ""
    echo "=========================================="
    echo "开始运行测试..."
    echo "=========================================="

    # 备份 input 目录
    TEST_BACKUP_DIR="test/input_backup_$(date +%s)"
    echo "备份 input 目录到 $TEST_BACKUP_DIR..."
    cp -r test/input "$TEST_BACKUP_DIR"

    # 运行程序处理 input
    echo ""
    echo "使用编译的程序处理 input 目录..."
    # 输入：test/input（指定目录），空行（执行1-3），3（退出）
    # 将输出重定向到 /dev/null 以保持输出清洁
    (echo "test/input"; echo ""; echo "3") | ./bin/after-photo-mac > /dev/null

    # 比较文件结构
    echo ""
    echo "比较文件结构..."

    # 检查所有文件是否都存在（排除 .DS_Store 文件和 .txt 日志文件）
    echo "检查结果..."
    OUTPUT_FILES=$(find test/output -type f -not -name '.DS_Store' -not -name '*.txt' | sort)
    INPUT_FILES=$(find test/input -type f -not -name '.DS_Store' -not -name '*.txt' | sort)

    # 获取文件名（去掉路径前缀）
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
    echo "编译完成（未运行测试）"
    echo "提示：使用 --test 或 -t 参数可以运行测试"
    echo "=========================================="
    exit 0
fi