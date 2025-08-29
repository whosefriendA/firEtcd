# firEtcd 演示项目 Makefile

.PHONY: help build start stop test clean demo

# 默认目标
help:
	@echo "🚀 firEtcd 服务发现演示项目"
	@echo ""
	@echo "可用命令:"
	@echo "  make build      - 构建 firEtcd 二进制文件"
	@echo "  make start      - 启动 firEtcd 集群"
	@echo "  make stop       - 停止 firEtcd 集群"
	@echo "  make test       - 运行端到端测试"
	@echo "  make demo       - 启动完整演示"
	@echo "  make clean      - 清理演示环境"
	@echo "  make health     - 检查集群健康状态"
	@echo ""

# 构建 firEtcd
build:
	@echo "🔨 构建 firEtcd..."
	cd .. && make build
	@echo "✅ 构建完成"

# 启动集群
start: build
	@echo "🚀 启动 firEtcd 集群..."
	@chmod +x scripts/*.sh
	./scripts/start-cluster.sh
	@echo "✅ 集群启动完成"

# 停止集群
stop:
	@echo "🛑 停止 firEtcd 集群..."
	./scripts/stop-cluster.sh
	@echo "✅ 集群已停止"

# 健康检查
health:
	@echo "🏥 检查集群健康状态..."
	./scripts/health-check.sh

# 运行端到端测试
test: start
	@echo "🧪 运行端到端测试..."
	@echo "等待集群稳定..."
	@sleep 10
	go test ./tests/ -v -run "TestE2E"
	@echo "✅ 测试完成"

# 运行性能测试
test-performance: start
	@echo "📊 运行性能测试..."
	@echo "等待集群稳定..."
	@sleep 10
	go test ./tests/ -v -run "TestE2EPerformance"
	@echo "✅ 性能测试完成"

# 启动完整演示
demo: start
	@echo "🎬 启动完整演示..."
	@echo ""
	@echo "📋 演示步骤:"
	@echo "1. 集群已启动，等待稳定..."
	@sleep 5
	@echo "2. 启动用户服务 (端口 8083)..."
	@echo "   新开终端运行: go run ./services/user-service/main.go"
	@echo ""
	@echo "3. 启动订单服务 (端口 8084)..."
	@echo "   新开终端运行: go run ./services/order-service/main.go"
	@echo ""
	@echo "4. 启动 API 网关 (端口 8080)..."
	@echo "   新开终端运行: go run ./services/gateway-service/main.go"
	@echo ""
	@echo "5. 运行客户端演示..."
	@echo "   新开终端运行: go run ./client/service-client.go"
	@echo ""
	@echo "6. 测试 HTTP API..."
	@echo "   健康检查: curl http://localhost:8080/health"
	@echo "   服务状态: curl http://localhost:8080/status"
	@echo "   服务发现: curl http://localhost:8080/api/discovery"
	@echo ""
	@echo "🎯 演示已准备就绪！"

# 启动单个服务
user-service: start
	@echo "👤 启动用户服务..."
	go run ./services/user-service/main.go

order-service: start
	@echo "📦 启动订单服务..."
	go run ./services/order-service/main.go

gateway: start
	@echo "🌐 启动 API 网关..."
	go run ./services/gateway-service/main.go

client-demo: start
	@echo "🎯 启动客户端演示..."
	@sleep 5
	go run ./client/service-client.go

# 清理演示环境
clean:
	@echo "🧹 清理演示环境..."
	./scripts/stop-cluster.sh
	rm -rf logs data
	@echo "✅ 清理完成"

# 快速演示（自动化）
quick-demo: start
	@echo "⚡ 快速演示模式..."
	@echo "等待集群启动..."
	@sleep 10
	@echo "运行端到端测试..."
	go test ./tests/ -v -run "TestE2E" -timeout 2m
	@echo "✅ 快速演示完成"

# 监控模式
monitor: start
	@echo "👀 监控模式启动..."
	@echo "集群已启动，开始监控..."
	@echo "按 Ctrl+C 停止监控"
	@trap 'make stop' INT; \
	while true; do \
		clear; \
		echo "=== firEtcd 集群监控 ==="; \
		echo "时间: $$(date)"; \
		echo ""; \
		./scripts/health-check.sh; \
		echo ""; \
		echo "按 Ctrl+C 停止监控"; \
		sleep 10; \
	done 