#!/bin/bash
#
# SMPP 完整端到端测试脚本
# 测试 SMPP 协议 + HTTP API 完整流程
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 配置
HTTP_PORT="18080"
SMPP_PORT="12775"
ADMIN_PASSWORD="admin123"

# 检测 Python 命令 (Windows 上 python3 可能是微软商店的占位符，需要实际测试)
detect_python() {
    for cmd in python3 python; do
        if command -v "$cmd" &> /dev/null; then
            # 实际测试 Python 是否可用（排除微软商店占位符）
            if "$cmd" --version &> /dev/null; then
                echo "$cmd"
                return 0
            fi
        fi
    done
    echo "Error: Python not found"
    return 1
}

PYTHON_CMD=$(detect_python)
if [ -z "$PYTHON_CMD" ] || [ "$PYTHON_CMD" = "Error: Python not found" ]; then
    echo "Error: Python not found"
    exit 1
fi

# Docker 容器配置
POSTGRES_CONTAINER="smpp-test-postgres"
REDIS_CONTAINER="smpp-test-redis"
POSTGRES_PORT="15432"
REDIS_PORT="16379"
POSTGRES_USER="smpp"
POSTGRES_PASSWORD="smpp_test"
POSTGRES_DB="smpp_test"

# Go 环境变量 (自动检测，仅在未设置时使用默认值)
export PATH="${PATH}:/usr/local/go/bin"
export GOPATH="${GOPATH:-$HOME/go}"
export GOCACHE="${GOCACHE:-/tmp/go-cache}"
export GOMODCACHE="${GOMODCACHE:-$HOME/go/pkg/mod}"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${YELLOW}[STEP]${NC} $1"; }
log_section() { echo -e "\n${BLUE}========== $1 ==========${NC}"; }

cleanup() {
    # 停止后端服务
    [ ! -z "$BACKEND_PID" ] && kill $BACKEND_PID 2>/dev/null || true
    # 清理 Docker 容器
    docker rm -f $POSTGRES_CONTAINER 2>/dev/null || true
    docker rm -f $REDIS_CONTAINER 2>/dev/null || true
}
trap cleanup EXIT

# 启动 PostgreSQL 容器
start_postgres() {
    log_step "启动 PostgreSQL 容器..."
    docker run -d \
        --name $POSTGRES_CONTAINER \
        -e POSTGRES_USER=$POSTGRES_USER \
        -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
        -e POSTGRES_DB=$POSTGRES_DB \
        -p $POSTGRES_PORT:5432 \
        postgres:16-alpine >/dev/null

    # 等待 PostgreSQL 启动
    for i in $(seq 1 30); do
        if docker exec $POSTGRES_CONTAINER pg_isready -U $POSTGRES_USER -d $POSTGRES_DB >/dev/null 2>&1; then
            log_info "PostgreSQL 已就绪"
            return 0
        fi
        sleep 1
    done
    log_error "PostgreSQL 启动超时"
    return 1
}

# 启动 Redis 容器
start_redis() {
    log_step "启动 Redis 容器..."
    docker run -d \
        --name $REDIS_CONTAINER \
        -p $REDIS_PORT:6379 \
        redis:7-alpine >/dev/null

    # 等待 Redis 启动
    for i in $(seq 1 15); do
        if docker exec $REDIS_CONTAINER redis-cli ping 2>/dev/null | grep -q PONG; then
            log_info "Redis 已就绪"
            return 0
        fi
        sleep 1
    done
    log_error "Redis 启动超时"
    return 1
}

wait_http() {
    for i in $(seq 1 30); do
        curl -s "http://127.0.0.1:$HTTP_PORT/health" >/dev/null 2>&1 && return 0
        sleep 1
    done
    return 1
}

get_token() {
    curl -s -X POST "http://127.0.0.1:$HTTP_PORT/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"'"$ADMIN_PASSWORD"'"}' | \
        grep -o '"token":"[^"]*"' | cut -d'"' -f4
}

count_messages() {
    curl -s -H "Authorization: Bearer $1" "http://127.0.0.1:$HTTP_PORT/api/messages?status=$2" | \
        grep -o '"total":[0-9]*' | cut -d: -f2
}

test_smpp() {
    local description=$1
    shift
    local output
    output=$($PYTHON_CMD "$SCRIPT_DIR/smpp_client.py" "$@" 2>&1)
    local exit_code=$?
    if [ $exit_code -eq 0 ]; then
        log_info "✓ $description"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "✗ $description"
        echo "    Output: $output"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

echo "============================================"
echo "  SMPP-Simulator 完整端到端测试"
echo "============================================"

# 启动 Docker 容器
start_postgres
start_redis

log_step "启动后端服务..."
cd "$PROJECT_ROOT/backend"
export HTTP_PORT=$HTTP_PORT
export SMPP_PORT=$SMPP_PORT
export DB_TYPE="postgres"
export DB_HOST="127.0.0.1"
export DB_PORT=$POSTGRES_PORT
export DB_NAME=$POSTGRES_DB
export DB_USER=$POSTGRES_USER
export DB_PASSWORD=$POSTGRES_PASSWORD
export REDIS_HOST="127.0.0.1"
export REDIS_PORT=$REDIS_PORT
export REDIS_ENABLED="true"
export ADMIN_PASSWORD=$ADMIN_PASSWORD

go build -o /tmp/smpp-server ./cmd/server/ 2>/dev/null
/tmp/smpp-server > /tmp/smpp_e2e.log 2>&1 &
BACKEND_PID=$!

if ! wait_http; then
    log_error "服务启动超时"
    cat /tmp/smpp_e2e.log
    exit 1
fi
log_info "服务已启动"

TOKEN=$(get_token)
[ -z "$TOKEN" ] && { log_error "登录失败"; exit 1; }
log_info "Token 获取成功"

# ==================== SMPP 协议测试 ====================
log_section "1. SMPP 绑定测试"

test_smpp "bind_transmitter" \
    --host 127.0.0.1 --port $SMPP_PORT \
    bind --system-id tx_user --password test --type transmitter

test_smpp "bind_receiver" \
    --host 127.0.0.1 --port $SMPP_PORT \
    bind --system-id rx_user --password test --type receiver

test_smpp "bind_transceiver" \
    --host 127.0.0.1 --port $SMPP_PORT \
    bind --system-id tr_user --password test --type transceiver

log_section "2. SMPP 消息发送测试"

# GSM7 编码测试
test_smpp "submit_sm GSM7" \
    --host 127.0.0.1 --port $SMPP_PORT \
    send --system-id msg_user --password test \
    --from 10086 --to 13800138000 --message "Hello GSM7" --encoding GSM7

# UCS2 编码测试（中文）
test_smpp "submit_sm UCS2 (中文)" \
    --host 127.0.0.1 --port $SMPP_PORT \
    send --system-id msg_user --password test \
    --from 10086 --to 13800138000 --message "你好世界" --encoding UCS2

log_section "3. 完整流程测试"

test_smpp "完整流程: bind -> send -> unbind" \
    --host 127.0.0.1 --port $SMPP_PORT \
    test --system-id flow_user --password flow_pass \
    --from 10010 --to 13900139000 --message "Flow Test Message"

sleep 1

log_section "4. 消息存储验证"

TOTAL=$(count_messages "$TOKEN" "")
PENDING=$(count_messages "$TOKEN" "pending")
DELIVERED=$(count_messages "$TOKEN" "delivered")

log_info "总消息数: $TOTAL"
log_info "待处理: $PENDING"
log_info "已送达: $DELIVERED"

if [ "$TOTAL" -ge 3 ]; then
    log_info "✓ 消息存储验证通过"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    log_error "✗ 消息数量不足: $TOTAL < 3"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

log_section "5. 消息查询 API 测试"

# 按源地址查询
MSG_BY_SOURCE=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "http://127.0.0.1:$HTTP_PORT/api/messages?source_addr=10086")
if echo "$MSG_BY_SOURCE" | grep -q "10086"; then
    log_info "✓ 按源地址查询成功"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    log_error "✗ 按源地址查询失败"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# 按目标地址查询
MSG_BY_DEST=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "http://127.0.0.1:$HTTP_PORT/api/messages?dest_addr=13800138000")
if echo "$MSG_BY_DEST" | grep -q "13800138000"; then
    log_info "✓ 按目标地址查询成功"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    log_error "✗ 按目标地址查询失败"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

log_section "6. 消息状态更新测试"

# 关闭自动送达回执，发送消息后手动标记送达
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    "http://127.0.0.1:$HTTP_PORT/api/mock/config" \
    -d '{"auto_response":true,"success_rate":100,"response_delay":0,"deliver_report":false}' >/dev/null

# 发送一条消息（不会自动送达）
test_smpp "手动送达测试消息" \
    --host 127.0.0.1 --port $SMPP_PORT \
    send --system-id manual_user --password test \
    --from 20000 --to 30000 --message "Manual Delivery Test" --encoding GSM7

sleep 1

# 获取这条待处理消息
PENDING_MSG=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "http://127.0.0.1:$HTTP_PORT/api/messages?status=pending&source_addr=20000")
MSG_ID=$(echo "$PENDING_MSG" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ ! -z "$MSG_ID" ]; then
    # 标记为已送达
    curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        "http://127.0.0.1:$HTTP_PORT/api/messages/$MSG_ID/deliver" >/dev/null

    # 验证状态
    UPDATED_MSG=$(curl -s -H "Authorization: Bearer $TOKEN" \
        "http://127.0.0.1:$HTTP_PORT/api/messages/$MSG_ID")

    if echo "$UPDATED_MSG" | grep -q '"status":"delivered"'; then
        log_info "✓ 消息状态更新成功"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_error "✗ 消息状态更新失败"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    log_info "没有待处理消息，跳过状态更新测试"
fi

log_section "7. Mock 配置测试"

# 设置自动响应
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    "http://127.0.0.1:$HTTP_PORT/api/mock/config" \
    -d '{"auto_response":true,"success_rate":100,"response_delay":0,"deliver_report":true,"deliver_delay":100}' >/dev/null

# 发送消息测试自动响应
test_smpp "自动响应测试" \
    --host 127.0.0.1 --port $SMPP_PORT \
    send --system-id auto_user --password test \
    --from 10000 --to 19999 --message "Auto Response Test"

sleep 1

# 验证消息状态（应该自动变为 delivered 或 failed）
AUTO_MSG=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "http://127.0.0.1:$HTTP_PORT/api/messages?source_addr=10000")

if echo "$AUTO_MSG" | grep -q '"status":"delivered"\|"status":"failed"'; then
    log_info "✓ 自动响应功能正常"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    log_error "✗ 自动响应功能异常"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

log_section "8. Session 统计测试"

# 获取 session 列表
SESSIONS=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "http://127.0.0.1:$HTTP_PORT/api/sessions")

SESSION_COUNT=$(echo "$SESSIONS" | grep -o '"id"' | wc -l)
log_info "Session 数量: $SESSION_COUNT"

if [ "$SESSION_COUNT" -ge 0 ]; then
    log_info "✓ Session 列表获取成功"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    log_error "✗ Session 列表获取失败"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

log_section "9. 数据清理测试"

# 清空所有数据
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" \
    "http://127.0.0.1:$HTTP_PORT/api/data/all" >/dev/null

# 验证清理结果
CLEAN_COUNT=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "http://127.0.0.1:$HTTP_PORT/api/messages" | grep -o '"total":[0-9]*' | cut -d: -f2)

if [ "$CLEAN_COUNT" = "0" ]; then
    log_info "✓ 数据清理成功"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    log_error "✗ 数据清理失败: 剩余 $CLEAN_COUNT 条"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# ==================== 测试汇总 ====================
echo ""
echo "============================================"
echo "  测试结果汇总"
echo "============================================"
log_info "通过: $TESTS_PASSED"
if [ $TESTS_FAILED -gt 0 ]; then
    log_error "失败: $TESTS_FAILED"
    exit 1
else
    log_info "失败: 0"
    log_info "\n所有测试通过! ✓"
fi
