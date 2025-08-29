#!/bin/bash

# firEtcd 集群停止脚本
set -e

echo "🛑 停止 firEtcd 集群..."

# 停止所有 firEtcd 进程
if pgrep -f "firEtcd" > /dev/null; then
    echo "📡 正在停止 firEtcd 进程..."
    pkill -f "firEtcd"
    
    # 等待进程完全停止
    sleep 3
    
    # 强制停止仍在运行的进程
    if pgrep -f "firEtcd" > /dev/null; then
        echo "⚠️  强制停止剩余进程..."
        pkill -9 -f "firEtcd"
    fi
else
    echo "ℹ️  没有发现运行中的 firEtcd 进程"
fi

# 清理 PID 文件
if [ -d "data" ]; then
    rm -f data/*.pid
    echo "🧹 清理 PID 文件完成"
fi

echo "✅ 集群已停止" 