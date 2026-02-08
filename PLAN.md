🚀 35天 Golang 后端架构师计划 

🎯 核心目标
项目: Flux-KV (分布式 KV 存储 + 高性能微服务网关 + CDC 实时数据流架构)

技术栈: Golang (GMP/Channel), gRPC, Etcd, RabbitMQ, Docker, OpenTelemetry

🗺️ 阶段一：分布式存储引擎 (Day 1-10)

[x] Day 1: 线程安全 Map (sync.RWMutex) + 竞态检测 ✅

[x] Day 2: TTL 过期清理机制 (Lazy + Active GC) ✅

[x] Day 3: AOF 持久化 (文件 IO + 启动恢复) ✅

[x] Day 4: TCP 服务端搭建 (自定义协议解决粘包) ✅

[x] Day 5: 客户端 SDK 封装 (连接池) ✅

[x] Day 6: gRPC 改造 (Protobuf 定义) ✅

[x] Day 7: gRPC 双向通信调试 & CLI ✅

[x] Day 8: 一致性哈希算法 (核心分片逻辑) ✅

[x] Day 9: [服务发现] 集成 Etcd 实现节点自动注册与发现 —— （大厂标准方案，替代手写 Gossip，更贴近实习需求）✅

[x] Day 10: 阶段总结 & Benchmark 压测 (基准测试) —— (完成 4.6万 QPS 压测与混沌工程实验，验证服务发现自愈能力) ✅

🗺️ 阶段二：高性能 API 网关 (Day 11-20)

[x] Day 11: HTTP Server 搭建 (Gin 框架集成) ✅

[x] Day 12: [泛化调用] HTTP 转 gRPC 动态代理 —— 网关的核心，让前端能调后端 ✅

[x] Day 13: [链路追踪] 集成 OpenTelemetry/Jaeger —— 微服务可观测性，大厂必问 ✅

[x] Day 14: [中间件] Access Log 异步日志库开发 (Channel 缓冲写) ✅

[x] Day 15: [高可用] 全局限流 (Token Bucket 算法实现) ✅

[x] Day 16: [高可用] 熔断降级 (Hystrix 状态机实现) ✅

[x] Day 17: [负载均衡] 结合 Day 9 的 Etcd 实现 RoundRobin 策略 (Discovery Watch + Client Pool) ✅

[x] Day 18: [防击穿] SingleFlight 模式实现 —— 高并发杀手锏 ✅

[x] Day 19: [性能调优] Pprof 性能分析与内存优化实战 ✅

[x] Day 20: 阶段复盘 & 压力测试 (wrk -> 52k QPS) ✅

🗺️ 阶段三：工程化完善与架构演进 (Day 21-30) 🔥 核心升级

[x] Day 21: [存储引擎] 锁竞争优化 (Sharded Map) —— 手写 FNV 哈希分片锁替代 sync.Map，大幅降低高并发写场景下的锁粒度冲突。 ✅

[x] Day 22: [中间件] RabbitMQ 与事件总线 (EventBus) —— 搭建基础设置，在 KV 内部实现基于 Channel 的 EventBus 解耦业务逻辑与消息发送。 ✅

[x] Day 23: [架构模式] CDC 数据变更流 (Change Data Capture) —— 改造 Set/Del 操作，在落盘后异步分发变更事件，实现 "Fire-and-Forget" 模式，确保主流程低延迟。 ✅

[x] Day 24: [可靠性] 优雅启停 (Graceful Shutdown) —— 完善 KV Server 端的退出逻辑，确保 AOF 缓冲区刷盘、RabbitMQ 连接安全关闭，防止数据丢失。 ✅

[ ] Day 25: [容器化] Docker Compose 集群编排 —— 编写多阶段构建 Dockerfile，一键拉起 Etcd, RabbitMQ, Jaeger, KV-Nodes(x3), Gateway 完整环境。

[ ] Day 26: [极致压测] 性能对比验证 —— 使用 pprof 对比 ShardedMap 优化前后的锁等待时间 (Mutex Wait)；验证 CDC 开启对写性能的影响。

[ ] Day 27: [代码重构] Go Idiomatic Refactoring —— 全局代码审查，优化 Error Handling，规范 Context 传递，清理硬编码与冗余逻辑。

[ ] Day 28: [项目门面] 架构文档与 README 完善 —— 绘制清晰的 "分布式架构图" (展示 CDC 流程) 和 "时序图"，更新 API 文档。

[ ] Day 29: [面试演练] 分布式系统专题 —— 准备 "强一致性 vs 最终一致性"、"CDC 的应用场景"、"为什么不用 MySQL Binlog" 等话术。

[ ] Day 30: [面试演练] Go 语言核心专题 —— 准备 "Map 底层扩容机制"、"G-M-P 调度模型"、"Channel 无锁编程原理" 等底层原理。

📝 阶段四：简历与面试 (Day 31-35)
（最后冲刺，面向面试）

[ ] Day 31: 简历深度优化 —— 将 "Sharded Map" 和 "CDC 架构" 作为核心亮点重写项目描述，突出 "高并发" 和 "解耦" 关键词。

[ ] Day 32: 模拟面试 (计算机基础) —— 计网 (TCP/HTTP/gRPC)、操作系统 (IO模型/进程线程)、数据库 (索引/事务/Redis)。

[ ] Day 33: 模拟面试 (项目设计) —— 能够白板画出项目整体架构，解释每个组件选型理由 (Etcd vs Consul, gRPC vs HTTP)。

[ ] Day 34: 模拟面试 (现场 Debug) —— 复盘之前的 Chaos Test 和 Pprof 调优过程，准备好讲述 "如何排查线上 OOM 或 CPU 飙升" 的故事。

[ ] Day 35: 最终复查与投递 —— 整理 Github 仓库，确保代码无敏感信息，Run 起来无报错，开始投递简历。