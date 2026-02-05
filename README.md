# Go-AI-KV-System

这是一个高性能的分布式 KV 存储系统与微服务网关项目，集成了 Go 语言核心特性与云原生技术栈。

本项目是 [35天 Golang 后端 + AI Agent 全栈突击计划](PLAN.md) 的实战成果。

## 🚀 功能特性

### 分布式 KV 存储
- **高性能存储**: 基于 sync.RWMutex 的线程安全 Map，支持 KV 读写。
- **持久化**: 支持 AOF (Append Only File) 持久化与启动恢复。
- **过期机制**: 实现 Lazy + Active 混合过期清理策略。
- **通信协议**: 自定义 TCP 协议（解决粘包问题）与 gRPC 接口支持。
- **一致性**: 一致性哈希算法实现数据分片。

### 微服务网关
- **服务发现**: 集成 Etcd 实现动态服务注册与发现。
- **动态代理**: HTTP 转 gRPC 泛化调用。
- **高可用**: 
  - 全局限流 (Token Bucket)
  - 熔断降级 (Hystrix)
  - 负载均衡 (RoundRobin)
  - 防缓存击穿 (SingleFlight)
- **可观测性**: 集成 OpenTelemetry/Jaeger 链路追踪与 Prometheus 指标监控。

## 🛠️ 快速开始

### 前置条件
- Go 1.22+
- Etcd (用于服务发现)
- Jaeger (可选，用于链路追踪)

### 运行服务端 (KV Server)
```bash
go run cmd/server/main.go
```

### 运行网关 (Gateway)
```bash
go run cmd/gateway/main.go
```

### 运行客户端测试
```bash
go run cmd/client/main.go
```

## 📚 学习计划
详细的开发日志和每日任务列表，请参阅 [PLAN.md](PLAN.md)。

## 📝 目录结构
```
├── api/            # IDL 定义 (Proto/gRPC)
├── cmd/            # 程序入口 (Gateway, Server, Client)
├── configs/        # 配置文件
├── internal/       # 私有业务逻辑
│   ├── core/       #存储引擎核心 (MemDB)
│   ├── gateway/    # 网关核心逻辑
│   ├── protocol/   # 通信协议
│   └── service/    # gRPC 服务实现
├── pkg/            # 公共库 (Client, Logger, Discovery)
├── scripts/        # 测试与运维脚本
└── tools/          # 工具集
```
