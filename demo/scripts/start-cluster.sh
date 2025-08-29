#!/bin/bash

# firEtcd 集群启动脚本
set -e

echo "🚀 启动 firEtcd 集群..."

# 创建必要的目录
mkdir -p logs
mkdir -p data

# 检查是否已经有进程在运行
if pgrep -f "firEtcd" > /dev/null; then
    echo "⚠️  发现已有 firEtcd 进程在运行，正在停止..."
    pkill -f "firEtcd"
    sleep 2
fi

# 启动第一个节点 (Leader)
echo "📡 启动节点 1 (Leader)..."
../bin/etcd/etcd -c=../cluster_config/node1/config.yml > logs/node1.log 2>&1 &
NODE1_PID=$!
echo $NODE1_PID > data/node1.pid

# 等待第一个节点启动
sleep 3

# 启动第二个节点
echo "📡 启动节点 2..."
../bin/etcd/etcd -c=../cluster_config/node2/config.yml > logs/node2.log 2>&1 &
NODE2_PID=$!
echo $NODE2_PID > data/node2.pid

# 启动第三个节点
echo "📡 启动节点 3..."
../bin/etcd/etcd -c=../cluster_config/node3/config.yml > logs/node3.log 2>&1 &
NODE3_PID=$!
echo $NODE3_PID > data/node3.pid

# 等待所有节点启动
echo "⏳ 等待集群启动..."
sleep 5

# 检查集群状态
echo "🔍 检查集群状态..."
echo "📊 当前运行的 etcd 进程："
ps aux | grep etcd | grep -v grep || echo "没有找到 etcd 进程"

if pgrep -f "etcd" > /dev/null; then
    echo "✅ 集群启动成功！"
    echo "📊 集群信息："
    echo "   - 节点 1: 端口 51240 (Raft: 32300)"
    echo "   - 节点 2: 端口 51241 (Raft: 32301)" 
    echo "   - 节点 3: 端口 51242 (Raft: 32302)"
    echo "   - Raft 端口: 32300, 32301, 32302"
    echo ""
    echo "📝 查看日志："
    echo "   - 节点 1: tail -f logs/node1.log"
    echo "   - 节点 2: tail -f logs/node2.log"
    echo "   - 节点 3: tail -f logs/node3.log"
else
    echo "❌ 集群启动失败，请检查日志"
    echo "📝 查看日志："
    echo "   - 节点 1: cat logs/node1.log"
    echo "   - 节点 2: cat logs/node2.log"
    echo "   - 节点 3: cat logs/node3.log"
    exit 1
fi

echo ""
echo "🎯 现在可以运行演示了："
echo "   go test ./tests/ -v -run 'TestE2E'" 