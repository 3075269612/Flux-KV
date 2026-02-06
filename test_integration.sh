#!/bin/bash
cd /home/wang/workspace/Flux-KV

echo "检查服务器状态..."
ps aux | grep "go run cmd/server" | grep -v grep

echo -e "\n准备客户端测试..."
sleep 2

echo -e "\n开始客户端集成测试..."
timeout 10 go run cmd/client/main.go <<EOF
set key1 value1
get key1
set user_1 tom
get user_1
del user_1
quit
EOF
