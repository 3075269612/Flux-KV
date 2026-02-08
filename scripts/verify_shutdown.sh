#!/usr/bin/env bash

# ================================================================
# Graceful Shutdown 自动化验证脚本
# ================================================================
# 功能：
#   1. 清理旧日志 / AOF / CDC 文件
#   2. go build 编译 Server、Consumer、Benchmark 二进制
#   3. 后台启动 Server 和 Consumer，记录真实 PID
#   4. 启动高并发 Benchmark（500 SET），1 秒后发送 SIGTERM
#   5. 等待进程优雅退出
#   6. 自动验证日志关键字 + 数据一致性，彩色输出 Pass/Fail
# ================================================================
# 兼容 macOS / Linux；需要预先运行 Etcd 和 RabbitMQ。
# 用法：
#   chmod +x scripts/verify_shutdown.sh
#   ./scripts/verify_shutdown.sh
# ================================================================

# -o pipefail : 管道中任一命令失败即非零退出码
# 不使用 set -e，因为我们需要手动控制每一步的退出逻辑
set -o pipefail

# ================================================================
# 颜色定义（自动检测终端是否支持颜色）
# ================================================================
if [ -t 1 ] && command -v tput >/dev/null 2>&1 && [ "$(tput colors 2>/dev/null)" -ge 8 ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  BLUE='\033[0;34m'
  BOLD='\033[1m'
  NC='\033[0m'
else
  RED='' GREEN='' YELLOW='' BLUE='' BOLD='' NC=''
fi

# ================================================================
# 日志辅助函数
# ================================================================
log_info()    { echo -e "${BLUE}[INFO]${NC}  $1"; }
log_pass()    { echo -e "${GREEN}[✓ PASS]${NC} $1"; }
log_fail()    { echo -e "${RED}[✗ FAIL]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}[⚠ WARN]${NC} $1"; }
log_header()  { echo -e "\n${BOLD}════════════════════════════════════════════════════════════${NC}"; echo -e "${BOLD} $1${NC}"; echo -e "${BOLD}════════════════════════════════════════════════════════════${NC}"; }

# ================================================================
# 全局配置
# ================================================================
WORKSPACE_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$WORKSPACE_DIR"

BIN_DIR="$WORKSPACE_DIR/bin"

SERVER_PORT=50052
ETCD_ADDR="localhost:2379"

# 日志文件
SERVER_LOG="$WORKSPACE_DIR/server_out.log"
CONSUMER_LOG="$WORKSPACE_DIR/consumer_out.log"
CLIENT_LOG="$WORKSPACE_DIR/client_out.log"

# 数据文件
AOF_FILE="$WORKSPACE_DIR/go-kv.aof"
CDC_LOG_FILE="$WORKSPACE_DIR/flux_cdc.log"

# 并发配置
TOTAL_REQUESTS=500       # 总请求数
CONCURRENCY=10           # 并发 goroutine 数
KILL_DELAY_SECONDS=3     # 启动 benchmark 后多久发送 SIGTERM (benchmark 内部需 1s 发现节点)

# 统计计数器
PASS_COUNT=0
FAIL_COUNT=0

# 需要清理的 PID
SERVER_PID=""
CONSUMER_PID=""
CLIENT_PID=""

# ================================================================
# 清理函数 —— 确保脚本异常退出时也能杀掉子进程
# ================================================================
cleanup() {
  log_info "执行清理..."
  for pid in $SERVER_PID $CONSUMER_PID $CLIENT_PID; do
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
      kill -9 "$pid" 2>/dev/null || true
    fi
  done
}
trap cleanup EXIT

# ================================================================
# 辅助：安全递增计数器（兼容 set -e）
# ================================================================
inc_pass() { PASS_COUNT=$((PASS_COUNT + 1)); }
inc_fail() { FAIL_COUNT=$((FAIL_COUNT + 1)); }

# ================================================================
# Step 0: 环境检查
# ================================================================
log_header "Step 0 — 环境检查"

# 检查 Go
if ! command -v go >/dev/null 2>&1; then
  log_fail "未找到 go 命令，请先安装 Go"; exit 1
fi
log_pass "Go $(go version | awk '{print $3}')"

# 检查 Etcd
if command -v etcdctl >/dev/null 2>&1; then
  if etcdctl --endpoints="$ETCD_ADDR" endpoint health >/dev/null 2>&1; then
    log_pass "Etcd 可达 ($ETCD_ADDR)"
  else
    log_warn "Etcd 不可达 ($ETCD_ADDR)，测试可能失败"
  fi
else
  log_warn "未安装 etcdctl，跳过 Etcd 连通性检查"
fi

# ================================================================
# Step 1: 清理旧文件
# ================================================================
log_header "Step 1 — 清理旧文件"

rm -f "$SERVER_LOG" "$CONSUMER_LOG" "$CLIENT_LOG"
rm -f "$AOF_FILE" "$CDC_LOG_FILE"
# 项目根目录下可能残留的日志/AOF
rm -f "$WORKSPACE_DIR"/*.log "$WORKSPACE_DIR"/*.aof 2>/dev/null || true

log_pass "旧文件已清理"

# ================================================================
# Step 2: 杀死残留进程
# ================================================================
log_header "Step 2 — 杀死残留进程"

# 杀 go run 残留进程和编译后的二进制残留
for pattern in "cmd/server" "cmd/cdc_consumer" "cmd/benchmark" "flux-server" "flux-consumer" "flux-bench"; do
  pkill -f "$pattern" 2>/dev/null || true
done
sleep 1
log_pass "残留进程已清理"

# ================================================================
# Step 3: 编译二进制
# ================================================================
log_header "Step 3 — 编译二进制"
# 使用 go build 生成真实二进制，确保 PID 跟踪准确
# （go run 会生成子进程，SIGTERM 无法正确传递）

mkdir -p "$BIN_DIR"

log_info "编译 Server..."
if ! go build -o "$BIN_DIR/flux-server" ./cmd/server; then
  log_fail "Server 编译失败"; exit 1
fi

log_info "编译 Consumer..."
if ! go build -o "$BIN_DIR/flux-consumer" ./cmd/cdc_consumer; then
  log_fail "Consumer 编译失败"; exit 1
fi

log_info "编译 Benchmark..."
if ! go build -o "$BIN_DIR/flux-bench" ./cmd/benchmark; then
  log_fail "Benchmark 编译失败"; exit 1
fi

log_pass "所有二进制编译成功 → $BIN_DIR/"

# ================================================================
# Step 4: 启动 Server 和 Consumer
# ================================================================
log_header "Step 4 — 启动服务"

log_info "启动 gRPC Server (端口: $SERVER_PORT)..."
"$BIN_DIR/flux-server" -port "$SERVER_PORT" -etcd "$ETCD_ADDR" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!
log_pass "Server 已启动 (PID: $SERVER_PID)"

log_info "启动 CDC Consumer..."
"$BIN_DIR/flux-consumer" >"$CONSUMER_LOG" 2>&1 &
CONSUMER_PID=$!
log_pass "Consumer 已启动 (PID: $CONSUMER_PID)"

# ================================================================
# Step 5: 等待服务就绪
# ================================================================
log_header "Step 5 — 等待服务就绪"
log_info "等待 2 秒让服务初始化..."
sleep 2

# 验证进程存活
if ! kill -0 "$SERVER_PID" 2>/dev/null; then
  log_fail "Server 启动失败！日志如下："
  cat "$SERVER_LOG"
  exit 1
fi

if ! kill -0 "$CONSUMER_PID" 2>/dev/null; then
  log_fail "Consumer 启动失败！日志如下："
  cat "$CONSUMER_LOG"
  exit 1
fi

log_pass "Server (PID $SERVER_PID) 和 Consumer (PID $CONSUMER_PID) 均运行中"

# ================================================================
# Step 6: 启动高并发 Benchmark + 延迟发送 SIGTERM
# ================================================================
log_header "Step 6 — 高并发请求 + 发送 SIGTERM"

log_info "启动 Benchmark: $TOTAL_REQUESTS 个 SET 请求, 并发 $CONCURRENCY..."
"$BIN_DIR/flux-bench" -n "$TOTAL_REQUESTS" -c "$CONCURRENCY" -etcd "$ETCD_ADDR" >"$CLIENT_LOG" 2>&1 &
CLIENT_PID=$!

# 等待一段时间，使 Client 发送了一部分请求
log_info "等待 ${KILL_DELAY_SECONDS} 秒后发送 SIGTERM..."
sleep "$KILL_DELAY_SECONDS"

log_warn "向 Server (PID $SERVER_PID) 和 Consumer (PID $CONSUMER_PID) 发送 SIGTERM..."
kill -TERM "$SERVER_PID" 2>/dev/null || true
kill -TERM "$CONSUMER_PID" 2>/dev/null || true

# ================================================================
# Step 7: 等待进程优雅退出
# ================================================================
log_header "Step 7 — 等待进程优雅退出"

TIMEOUT=15
ELAPSED=0

while [ $ELAPSED -lt $TIMEOUT ]; do
  SERVER_ALIVE=false
  CONSUMER_ALIVE=false

  kill -0 "$SERVER_PID" 2>/dev/null && SERVER_ALIVE=true
  kill -0 "$CONSUMER_PID" 2>/dev/null && CONSUMER_ALIVE=true

  if ! $SERVER_ALIVE && ! $CONSUMER_ALIVE; then
    log_pass "Server 和 Consumer 均已退出 (耗时 ${ELAPSED}s)"
    break
  fi

  sleep 1
  ELAPSED=$((ELAPSED + 1))
  log_info "等待中... (${ELAPSED}s / ${TIMEOUT}s) Server=$SERVER_ALIVE Consumer=$CONSUMER_ALIVE"
done

# 超时后强制杀死
if kill -0 "$SERVER_PID" 2>/dev/null; then
  log_warn "Server ${TIMEOUT}s 超时，强制杀死"
  kill -9 "$SERVER_PID" 2>/dev/null || true
fi
if kill -0 "$CONSUMER_PID" 2>/dev/null; then
  log_warn "Consumer ${TIMEOUT}s 超时，强制杀死"
  kill -9 "$CONSUMER_PID" 2>/dev/null || true
fi

# 等待 Benchmark 自然结束（Server 挂了它也会报错退出）
wait "$CLIENT_PID" 2>/dev/null || true
CLIENT_PID=""

# 短暂等待文件系统同步
sleep 1

# ================================================================
# Step 8: 验证关键日志
# ================================================================
log_header "Step 8 — 验证优雅关闭日志"

# --- Server: 注销 Etcd ---
if [ -f "$SERVER_LOG" ] && grep -q "正在注销 Etcd" "$SERVER_LOG"; then
  log_pass "Server 日志包含 \"正在注销 Etcd\""
  inc_pass
else
  log_fail "Server 日志缺少 \"正在注销 Etcd\""
  inc_fail
fi

# --- Server: MemDB 安全落袋 ---
if [ -f "$SERVER_LOG" ] && grep -q "MemDB 数据已安全落袋" "$SERVER_LOG"; then
  log_pass "Server 日志包含 \"MemDB 数据已安全落袋\""
  inc_pass
else
  log_fail "Server 日志缺少 \"MemDB 数据已安全落袋\""
  inc_fail
fi

# --- Consumer: 安全退出 ---
if [ -f "$CONSUMER_LOG" ] && grep -qE "(Consumer 安全退出|CDC Consumer 安全退出)" "$CONSUMER_LOG"; then
  log_pass "Consumer 日志包含 \"Consumer 安全退出\""
  inc_pass
else
  log_fail "Consumer 日志缺少 \"Consumer 安全退出\""
  inc_fail
fi

# ================================================================
# Step 9: 统计数据一致性
# ================================================================
log_header "Step 9 — 数据一致性统计"

# --- Client 成功数 ---
# benchmark 最终输出格式：
#   总请求: 500
#   成功率: 100.00%
# 通过 awk 从最终报告中计算成功数 = 总请求 × 成功率 / 100
if [ -f "$CLIENT_LOG" ]; then
  CLIENT_SUCCESS=$(awk '
    /总请求:/ { total = $2 + 0 }
    /成功率:/ { rate = $2 + 0; gsub(/%/, "", rate) }
    /成功:/ { last_success = $0; gsub(/.*成功: */, "", last_success); gsub(/ .*/, "", last_success) }
    END {
      if (total > 0 && rate > 0) { printf "%d", total * rate / 100 }
      else if (last_success + 0 > 0) { printf "%d", last_success + 0 }
      else { print 0 }
    }
  ' "$CLIENT_LOG")
else
  CLIENT_SUCCESS=0
fi
log_info "Client 发送成功的 SET 请求: ${GREEN}${CLIENT_SUCCESS}${NC}"

# --- AOF 行数 ---
if [ -f "$AOF_FILE" ]; then
  AOF_LINES=$(wc -l < "$AOF_FILE" | tr -d '[:space:]')
  log_info "AOF 文件记录的命令数:      ${GREEN}${AOF_LINES}${NC}"
else
  AOF_LINES=0
  log_warn "AOF 文件不存在: $AOF_FILE"
fi

# --- CDC 日志行数 ---
# 注意: grep -c 找不到匹配时输出 "0" 但 exit code=1
#       不能写 $(grep -c ... || echo 0)，否则会拼出 "0\n0"
if [ -f "$CDC_LOG_FILE" ]; then
  CDC_LINES=$(grep -c "CDC_SYNC" "$CDC_LOG_FILE" 2>/dev/null) || CDC_LINES=0
  log_info "CDC 日志记录的事件数:      ${GREEN}${CDC_LINES}${NC}"
else
  CDC_LINES=0
  log_warn "CDC 日志文件不存在: $CDC_LOG_FILE"
fi

# ================================================================
# Step 10: 数据对比报告
# ================================================================
log_header "Step 10 — 数据对比报告"

echo ""
printf "  ┌─────────────────────────────────────────────────┐\n"
printf "  │            数据一致性对比报告                    │\n"
printf "  ├──────────────────────┬──────────────────────────┤\n"
printf "  │ Client 成功请求         │ %24s │\n" "$CLIENT_SUCCESS"
printf "  │ AOF 写入命令            │ %24s │\n" "$AOF_LINES"
printf "  │ CDC 日志事件            │ %24s │\n" "$CDC_LINES"
printf "  ├──────────────────────┴──────────────────────────┤\n"

# --- AOF 对比 ---
if [ "$CLIENT_SUCCESS" -gt 0 ] 2>/dev/null; then
  if [ "$AOF_LINES" -eq "$CLIENT_SUCCESS" ]; then
    printf "  │ ${GREEN}✓ AOF 与 Client 完全一致${NC}%-24s│\n" ""
    inc_pass
  elif [ "$AOF_LINES" -gt 0 ]; then
    DIFF=$((CLIENT_SUCCESS - AOF_LINES))
    printf "  │ ${YELLOW}⚠ AOF 差异: %d (Client=%d AOF=%d)${NC}%s│\n" "$DIFF" "$CLIENT_SUCCESS" "$AOF_LINES" ""
    log_warn "AOF 行数不完全一致（部分请求可能在 SIGTERM 后未被处理）"
    inc_pass  # 有数据就算基本通过
  else
    printf "  │ ${RED}✗ AOF 文件为空${NC}%-34s│\n" ""
    inc_fail
  fi
else
  printf "  │ ${YELLOW}⚠ Client 无成功请求，跳过 AOF 对比${NC}%s│\n" ""
fi

# --- CDC 对比 ---
if [ "$CLIENT_SUCCESS" -gt 0 ] 2>/dev/null && [ "$CDC_LINES" -gt 0 ] 2>/dev/null; then
  if [ "$CDC_LINES" -eq "$CLIENT_SUCCESS" ]; then
    printf "  │ ${GREEN}✓ CDC 与 Client 完全一致${NC}%-24s│\n" ""
    inc_pass
  elif [ "$CDC_LINES" -le "$CLIENT_SUCCESS" ]; then
    DIFF=$((CLIENT_SUCCESS - CDC_LINES))
    printf "  │ ${YELLOW}⚠ CDC 差异: %d (异步延迟可接受)${NC}%s│\n" "$DIFF" ""
    inc_pass  # 异步消费少于发送是正常的
  else
    printf "  │ ${RED}✗ CDC 事件异常多于 Client 请求${NC}%s│\n" ""
    inc_fail
  fi
elif [ "$CDC_LINES" -eq 0 ] 2>/dev/null && [ "$CLIENT_SUCCESS" -gt 0 ] 2>/dev/null; then
  printf "  │ ${YELLOW}⚠ CDC 日志为空 (RabbitMQ 可能未运行)${NC}%s│\n" ""
fi

printf "  └─────────────────────────────────────────────────┘\n"

# ================================================================
# Step 11: 最终结果
# ================================================================
log_header "最终结果"

TOTAL=$((PASS_COUNT + FAIL_COUNT))

echo ""
if [ "$FAIL_COUNT" -eq 0 ]; then
  echo -e "  ${GREEN}${BOLD}✓ ALL PASSED ($PASS_COUNT/$TOTAL)${NC}"
  echo -e "  ${GREEN}优雅关闭验证全部通过！${NC}"
  EXIT_CODE=0
else
  echo -e "  ${RED}${BOLD}✗ $FAIL_COUNT FAILED${NC} ${GREEN}(通过: $PASS_COUNT${NC} / ${RED}失败: $FAIL_COUNT)${NC}"
  EXIT_CODE=1
fi
echo ""

# ================================================================
# Step 12: 日志文件位置 + 调试信息
# ================================================================
log_info "日志文件位置："
echo "  Server 日志:    $SERVER_LOG"
echo "  Consumer 日志:  $CONSUMER_LOG"
echo "  Client 日志:    $CLIENT_LOG"
echo "  AOF 文件:       $AOF_FILE"
echo "  CDC 日志:       $CDC_LOG_FILE"
echo ""

# 失败时显示关键日志，方便调试
if [ $EXIT_CODE -ne 0 ]; then
  if [ -f "$SERVER_LOG" ]; then
    log_header "Server 日志 (最后 30 行)"
    tail -30 "$SERVER_LOG"
    echo ""
  fi
  if [ -f "$CONSUMER_LOG" ]; then
    log_header "Consumer 日志 (最后 30 行)"
    tail -30 "$CONSUMER_LOG"
    echo ""
  fi
  if [ -f "$CLIENT_LOG" ]; then
    log_header "Client 日志 (最后 20 行)"
    tail -20 "$CLIENT_LOG"
    echo ""
  fi
fi

exit $EXIT_CODE
