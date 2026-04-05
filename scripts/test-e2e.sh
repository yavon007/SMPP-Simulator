#!/bin/bash
#
# SMPP-Simulator 完整端到端测试脚本
# 覆盖所有 API 端点
#

set -e

# 配置
BACKEND_HOST="127.0.0.1"
HTTP_PORT="18080"
SMPP_PORT="12775"
ADMIN_PASSWORD="admin123"
TIMEOUT=30

# Go 环境变量
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
export GOCACHE=/tmp/go-cache
export GOMODCACHE=/root/go/pkg/mod

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 计数器
TESTS_PASSED=0
TESTS_FAILED=0

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_section() { echo -e "\n${BLUE}========== $1 ==========${NC}"; }

# 清理函数
cleanup() {
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
        wait $BACKEND_PID 2>/dev/null || true
    fi
    rm -f /tmp/smpp_test_*.db 2>/dev/null || true
    rm -f /tmp/smpp_export_* 2>/dev/null || true
}
trap cleanup EXIT

check_port() {
    if lsof -i:$1 >/dev/null 2>&1; then
        log_error "端口 $1 已被占用"
        exit 1
    fi
}

wait_for_server() {
    local url="http://127.0.0.1:$HTTP_PORT/health"
    local count=0
    while [ $count -lt $TIMEOUT ]; do
        if curl -s "$url" >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
        count=$((count + 1))
    done
    return 1
}

# API 测试函数
test_api() {
    local method=$1
    local path=$2
    local data=$3
    local expected_status=$4
    local token=$5
    local description=$6

    local url="http://127.0.0.1:$HTTP_PORT$path"
    local args="-s -w '\n%{http_code}'"

    if [ ! -z "$token" ]; then
        args="$args -H 'Authorization: Bearer $token'"
    fi

    if [ "$method" = "POST" ] || [ "$method" = "PUT" ]; then
        args="$args -X $method -H 'Content-Type: application/json' -d '$data'"
    elif [ "$method" = "DELETE" ]; then
        args="$args -X DELETE"
        if [ ! -z "$data" ]; then
            args="$args -H 'Content-Type: application/json' -d '$data'"
        fi
    fi

    local response
    response=$(eval curl $args "$url" 2>/dev/null)
    local status=$(echo "$response" | tail -1)
    local body=$(echo "$response" | sed '$d')

    if [ "$status" = "$expected_status" ]; then
        log_info "✓ $method $path -> $status ${description:+[$description]}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo "$body"
        return 0
    else
        log_error "✗ $method $path -> $status (期望 $expected_status) ${description:+[$description]}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo "$body"
        return 1
    fi
}

# 主测试流程
main() {
    echo "============================================"
    echo "  SMPP-Simulator 完整 E2E 测试"
    echo "============================================"

    cd /root/SMPP-Simulator/backend

    log_info "检查端口..."
    check_port $HTTP_PORT
    check_port $SMPP_PORT

    log_info "启动后端服务..."
    export HTTP_PORT=$HTTP_PORT
    export SMPP_PORT=$SMPP_PORT
    export DB_PATH="/tmp/smpp_test_$(date +%s).db"
    export ADMIN_PASSWORD=$ADMIN_PASSWORD

    go build -o /tmp/smpp-server ./cmd/server/ 2>/tmp/smpp_build.log
    /tmp/smpp-server > /tmp/smpp_test_backend.log 2>&1 &
    BACKEND_PID=$!

    log_info "等待服务启动..."
    if ! wait_for_server; then
        log_error "服务启动超时"
        cat /tmp/smpp_test_backend.log
        exit 1
    fi
    log_info "服务已启动\n"

    # ==================== 健康检查 ====================
    log_section "1. 健康检查"
    test_api "GET" "/health" "" "200" "" "健康检查"

    # ==================== 公开 API ====================
    log_section "2. 公开 API (无需认证)"
    test_api "GET" "/api/stats" "" "200" "" "统计数据"
    test_api "GET" "/api/messages" "" "200" "" "消息列表"
    test_api "GET" "/api/messages?status=pending" "" "200" "" "按状态筛选"
    test_api "GET" "/api/messages?source_addr=10086" "" "200" "" "按源地址筛选"

    # ==================== 认证 API ====================
    log_section "3. 认证 API"

    # 登录
    LOGIN_RESPONSE=$(test_api "POST" "/api/auth/login" '{"username":"admin","password":"'"$ADMIN_PASSWORD"'"}' "200" "" "正确密码登录")
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

    if [ -z "$TOKEN" ]; then
        log_error "登录失败，无法获取 token"
        exit 1
    fi
    log_info "Token: ${TOKEN:0:30}..."

    test_api "POST" "/api/auth/login" '{"username":"admin","password":"wrongpassword"}' "401" "" "错误密码"
    test_api "GET" "/api/auth/status" "" "200" "$TOKEN" "检查登录状态"

    # ==================== Session API ====================
    log_section "4. Session API"

    test_api "GET" "/api/sessions" "" "200" "$TOKEN" "Session 列表"

    # ==================== 消息 API ====================
    log_section "5. 消息 API"

    # 先通过 SMPP 发送一条消息
    python3 /root/SMPP-Simulator/scripts/smpp_client.py \
        --host 127.0.0.1 --port $SMPP_PORT \
        test --system-id test_user --password test --from 10086 --to 13800 --message "Test Message" >/dev/null 2>&1 || true

    sleep 1

    # 获取消息列表
    MESSAGES_RESPONSE=$(test_api "GET" "/api/messages" "" "200" "$TOKEN" "消息列表")
    MSG_ID=$(echo "$MESSAGES_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ ! -z "$MSG_ID" ]; then
        test_api "GET" "/api/messages/$MSG_ID" "" "200" "$TOKEN" "消息详情"
        test_api "POST" "/api/messages/$MSG_ID/deliver" "" "200" "$TOKEN" "标记已送达"
        test_api "POST" "/api/messages/$MSG_ID/fail" "" "200" "$TOKEN" "标记失败"

        # 获取另一条消息用于批量删除测试
        python3 /root/SMPP-Simulator/scripts/smpp_client.py \
            --host 127.0.0.1 --port $SMPP_PORT \
            test --system-id test_user --password test --from 10087 --to 13801 --message "Batch Test" >/dev/null 2>&1 || true
        sleep 1

        MESSAGES_RESPONSE=$(test_api "GET" "/api/messages" "" "200" "$TOKEN" "获取消息")
        MSG_IDS=$(echo "$MESSAGES_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 | tr '\n' ',' | sed 's/,$//')

        if [ ! -z "$MSG_IDS" ]; then
            test_api "DELETE" "/api/messages/batch" "{\"ids\":[\"${MSG_IDS//,/\",\"}\"]}" "200" "$TOKEN" "批量删除消息"
        fi
    fi

    # 导出测试
    test_api "GET" "/api/messages/export" "" "200" "$TOKEN" "导出消息(JSON)"
    test_api "GET" "/api/messages/export?format=csv" "" "200" "$TOKEN" "导出消息(CSV)"

    # ==================== 模板 API ====================
    log_section "6. 模板 API"

    test_api "GET" "/api/templates" "" "200" "$TOKEN" "模板列表"

    TEMPLATE_RESPONSE=$(test_api "POST" "/api/templates" '{"name":"测试模板","content":"Hello {name}"}' "201" "$TOKEN" "创建模板")
    TEMPLATE_ID=$(echo "$TEMPLATE_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ ! -z "$TEMPLATE_ID" ]; then
        test_api "GET" "/api/templates/$TEMPLATE_ID" "" "200" "$TOKEN" "获取模板"
        test_api "PUT" "/api/templates/$TEMPLATE_ID" '{"name":"更新模板","content":"Hi {name}"}' "200" "$TOKEN" "更新模板"
        test_api "DELETE" "/api/templates/$TEMPLATE_ID" "" "200" "$TOKEN" "删除模板"
    fi

    # ==================== Mock 配置 API ====================
    log_section "7. Mock 配置 API"

    test_api "GET" "/api/mock/config" "" "200" "$TOKEN" "获取配置"
    test_api "PUT" "/api/mock/config" '{"auto_response":true,"success_rate":90,"response_delay":100,"deliver_report":true,"deliver_delay":500}' "200" "$TOKEN" "更新配置"

    # ==================== Send Message API ====================
    log_section "8. 下发消息 API"

    test_api "GET" "/api/send/receivers" "" "200" "$TOKEN" "获取接收方列表"

    # 需要 SMPP session 才能发送，这里只测试 API 响应
    test_api "POST" "/api/send" '{"session_id":"non-existent","source_addr":"10086","dest_addr":"13800","content":"test"}' "400" "$TOKEN" "发送到不存在的session"

    # ==================== Outbound API ====================
    log_section "9. Outbound 主动连接 API"

    test_api "GET" "/api/outbound" "" "200" "$TOKEN" "列出主动连接"

    # 尝试连接到本地 SMPP 服务器
    CONNECT_RESPONSE=$(test_api "POST" "/api/outbound/connect" '{"host":"127.0.0.1","port":"'"$SMPP_PORT"'","system_id":"outbound_test","password":"test","bind_type":"transceiver"}' "200" "$TOKEN" "连接远程SMSC")
    OUTBOUND_ID=$(echo "$CONNECT_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ ! -z "$OUTBOUND_ID" ]; then
        test_api "POST" "/api/outbound/$OUTBOUND_ID/send" '{"source_addr":"10086","dest_addr":"13800138000","content":"Outbound Test"}' "200" "$TOKEN" "通过主动连接发送"
        test_api "DELETE" "/api/outbound/$OUTBOUND_ID" "" "200" "$TOKEN" "断开主动连接"
    fi

    # ==================== 系统配置 API ====================
    log_section "10. 系统配置 API"

    test_api "GET" "/api/system/config" "" "200" "$TOKEN" "获取系统配置"
    test_api "PUT" "/api/system/config" '{"jwt_expiry":48}' "200" "$TOKEN" "更新系统配置"
    test_api "GET" "/api/system/redis" "" "200" "$TOKEN" "检查 Redis 状态"

    # ==================== 限流状态 API ====================
    log_section "11. 限流状态 API"

    test_api "GET" "/api/stats/rate-limit" "" "200" "$TOKEN" "获取限流状态"

    # ==================== 数据清理 API ====================
    log_section "12. 数据清理 API"

    test_api "DELETE" "/api/data/messages" "" "200" "$TOKEN" "清空消息"
    test_api "DELETE" "/api/data/sessions" "" "200" "$TOKEN" "清空 Session"
    test_api "DELETE" "/api/data/all" "" "200" "$TOKEN" "清空所有数据"

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
}

main "$@"
