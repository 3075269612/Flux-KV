# Flux-KV 简历优化方案

## 一、优化后的专业技能

> 改动说明：
> 1. 删除了原第2条"微服务架构"整行，将 gRPC、Etcd 等具体技能合并到其他条目
> 2. 将 Pprof 性能调优从"工程与运维"中独立强调

• **编程基础**：熟练掌握 Golang 及其并发模型 (Goroutine/Channel/sync 包)，熟悉 C/C++ 与 Python，具备扎实的数据结构与算法功底。

• **存储与数据库**：深入理解 MySQL 索引与事务原理及 SQL 优化，熟悉 Redis 高并发缓存与 Milvus 向量检索；熟练使用 gRPC + Protobuf 构建高性能 RPC 通信。

• **中间件与治理**：掌握 RabbitMQ 在异步解耦与 CDC 中的应用，掌握基于 Etcd 的服务注册发现与一致性哈希路由；了解熔断、令牌桶限流及 SingleFlight 防缓存击穿机制。

• **性能分析与调优**：具备基于 Pprof CPU Profile + 火焰图的性能瓶颈定位经验，能通过对照实验量化优化效果。

• **工程与运维**：熟练掌握 Linux 环境与 Git 协作，熟悉 Docker 容器化部署与 Docker Compose 编排，了解 OpenTelemetry/Jaeger 链路追踪。

• **AI 领域探索**：熟悉 RAG 架构与 Agent 开发流程，了解 ReAct 模式及大模型工具调用编排。

---

## 二、优化后的项目描述

### 高性能分布式 KV 存储系统 (Flux-KV) &emsp; 2026.01 – 2026.02

**技术栈**：Go、gRPC、Protobuf、Etcd、RabbitMQ、Pprof、Docker Compose、Jaeger

**项目简介**：基于 Go 开发的分布式键值存储系统，实现服务发现、负载均衡、CDC 事件流等企业级特性。单节点 QPS 达 4.2 万，256 分片设计将锁竞争控制在极低水平 (pprof 验证锁等待不在 Top 耗时)。

**核心工作**：

• **并发优化**：设计 256 分片 ShardedMap + FNV-1a 哈希定位分片 + sync.RWMutex 细粒度并发控制，100 并发 50 万次写入耗时 11.82s，QPS 达 42,285。

• **持久化机制**：实现 AOF (Append Only File) 日志持久化，JSON 格式逐行追加写入，启动时重放恢复，恢复速度达 5 万条/秒。

• **分布式能力**：基于 Etcd 实现服务注册发现 (带 Lease 自动续约 + Watch 感知上下线)，结合一致性哈希算法 (CRC32 + 150 虚拟节点) 实现精准的路由负载均衡，支持节点动态扩缩容。

• **CDC 事件流**：集成 RabbitMQ 实现数据变更捕获 (Change Data Capture)，采用带缓冲 Channel (10000) 异步投递替代同步发布，在 Handler 层实现非阻塞事件分发，性能损耗控制在 23.5% (42k → 32k QPS)。

• **Pprof 性能分析**：通过 Pprof CPU Profile 对 CDC 开启/关闭两组场景进行对照压测，生成火焰图定位瓶颈——确认 ShardedMap 有效消除全局锁热点 (sync.Mutex 等待时间未进入 Top 耗时)，主要 CPU 开销集中在 syscall.Write (网络 IO) 和 runtime.mallocgc (内存分配)；据此采用带缓冲 Channel 异步投递方案优化 RabbitMQ 写入路径，将 CDC 性能损耗从同步模式降低至 23.5%。

• **服务治理与可观测性**：Gateway 层集成令牌桶限流 (golang.org/x/time/rate)、Hystrix 熔断降级、SingleFlight 防缓存击穿 (对同一 Key 的并发请求合并为一次后端调用)、OpenTelemetry + Jaeger 全链路追踪；支持 Docker Compose 一键部署全栈环境。

---

## 三、改动对比 (Diff)

### 专业技能改动点

```diff
- • 微服务架构：熟悉微服务体系，熟练使用 gRPC 通信，掌握基于 Etcd 的服务注册发现与路由负载均衡。

- • 数据库引擎：深入理解 MySQL 索引与事务原理及 SQL 优化，熟悉 Redis 高并发缓存与 Milvus 向量检索。
+ • 存储与数据库：深入理解 MySQL 索引与事务原理及 SQL 优化，熟悉 Redis 高并发缓存与 Milvus 向量检索；熟练使用 gRPC + Protobuf 构建高性能 RPC 通信。

- • 中间件与治理：掌握 RabbitMQ 在异步解耦与 CDC 中的应用，了解熔断、令牌桶限流及并发防击穿机制。
+ • 中间件与治理：掌握 RabbitMQ 在异步解耦与 CDC 中的应用，掌握基于 Etcd 的服务注册发现与一致性哈希路由；了解熔断、令牌桶限流及 SingleFlight 防缓存击穿机制。

- • 工程与运维：熟练掌握 Linux 环境与 Git 协作，熟悉 Docker 容器化部署，具备基于 Pprof 的性能调优经验。
+ • 性能分析与调优：具备基于 Pprof CPU Profile + 火焰图的性能瓶颈定位经验，能通过对照实验量化优化效果。
+ • 工程与运维：熟练掌握 Linux 环境与 Git 协作，熟悉 Docker 容器化部署与 Docker Compose 编排，了解 OpenTelemetry/Jaeger 链路追踪。
```

### 项目描述改动点

```diff
  技术栈：
- Go、gRPC、Protobuf、Etcd、RabbitMQ、Docker Compose、Jaeger
+ Go、gRPC、Protobuf、Etcd、RabbitMQ、Pprof、Docker Compose、Jaeger

  核心工作：
- • 并发优化: 设计 256 分片 ShardedMap + RWMutex 并发控制，50 万次写入耗时 11.82s，锁等待 < 5% (通过 pprof 性能分析验证)。
+ • 并发优化：设计 256 分片 ShardedMap + FNV-1a 哈希定位分片 + sync.RWMutex 细粒度并发控制，100 并发 50 万次写入耗时 11.82s，QPS 达 42,285。

- • CDC 事件流: 集成 RabbitMQ 实现数据变更流,性能损耗控制在 23.5% (42k→32k QPS)。
+ • CDC 事件流：集成 RabbitMQ 实现数据变更捕获 (CDC)，采用带缓冲 Channel 异步投递替代同步发布，性能损耗控制在 23.5% (42k → 32k QPS)。

+ • Pprof 性能分析 (新增独立要点)：完整描述工具→发现→优化→效果的分析流程。

- • 微服务治理: 引入熔断器、限流器(令牌桶算法)、Jaeger 链路追踪，Docker Compose 一键部署。
+ • 服务治理与可观测性：增加 SingleFlight 防缓存击穿描述，改名避免"微服务"大词。
```
