#!/bin/bash

# firEtcd 集群健康检查脚本
set -e

echo "🔍 firEtcd 集群健康检查..."

# 检查节点 1
echo "📡 检查节点 1 (http://localhost:8080)..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ 节点 1 健康"
else
    echo "❌ 节点 1 不健康"
fi

# 检查节点 2
echo "📡 检查节点 2 (http://localhost:8081)..."
if curl -s http://localhost:8081/health > /dev/null 2>&1; then
    echo "✅ 节点 2 健康"
else
    echo "❌ 节点 2 不健康"
fi

# 检查节点 3
echo "📡 检查节点 3 (http://localhost:8082)..."
if curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo "✅ 节点 3 健康"
else
    echo "❌ 节点 3 不健康"
fi

# 检查集群状态
echo "📊 检查集群状态..."
if curl -s http://localhost:8080/cluster/status > /dev/null 2>&1; then
    echo "✅ 集群状态正常"
else
    echo "❌ 集群状态异常"
fi

# 检查服务发现 API
echo "🔍 检查服务发现 API..."
if curl -s "http://localhost:8080/api/services?service_name=test" > /dev/null 2>&1; then
    echo "✅ 服务发现 API 正常"
else
    echo "❌ 服务发现 API 异常"
fi

echo ""
echo "🏥 健康检查完成" 