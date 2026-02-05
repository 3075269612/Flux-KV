🚀 35天 Golang 后端 + AI Agent 全栈突击计划 (最终纯净版)

🎯 核心目标
项目: 分布式 KV 存储 + 高性能微服务网关 + Go原生多智能体编排引擎

算法: LeetCode Hot 100 (每日 3 题)

技术栈: Golang (GMP/Channel), gRPC, Etcd, Docker, LLM (Function Calling)

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

[ ] Day 18: [防击穿] SingleFlight 模式实现 —— 高并发杀手锏

[ ] Day 19: [性能调优] Pprof 性能分析与内存优化实战

[ ] Day 20: 阶段复盘 & 压力测试 (wrk)

🗺️ 阶段三：AI Agent 编排引擎 (Day 21-30) 🔥 核心升级

[ ] Day 21: LLM SDK 封装 (适配 OpenAI/DeepSeek 接口)

[ ] Day 22: [并发基石] 定义 Agent 结构体与 Actor 模型 —— 每个 Agent 一个 Goroutine

[ ] Day 23: [工具层] 实现 Tool 接口与反射调用 —— 让 AI 能调用你的 KV Store

[ ] Day 24: [大脑层] 实现 Function Calling 解析器 —— 处理 AI 返回的 JSON 指令

[ ] Day 25: [编排层] 实现 Supervisor (主管) Agent —— 负责拆解用户任务

[ ] Day 26: [异步解耦] 集成 RabbitMQ 实现 Worker 任务队列 —— 解决 LLM 响应慢难题

[ ] Day 27: [循环机制] 实现思考-执行-观察 Loop (ReAct 模式)

[ ] Day 28: [记忆集成] 将 KV Store 接入作为 Agent 的长期记忆 (Session Storage)

[ ] Day 29: Docker 容器化与 Compose 编排

[ ] Day 30: 系统联调：用自然语言指令控制集群 (End-to-End Demo)

📝 阶段四：简历与面试 (Day 31-35)
（最后冲刺，面向面试）

[ ] Day 31: 架构图绘制 (突出 Swarm 并发模型)

[ ] Day 32: 难点深挖 (Golang GMP 调度在 Agent 中的应用)

[ ] Day 33: 模拟面试 (计算机基础：计网/OS/DB)

[ ] Day 34: 模拟面试 (项目篇：微服务 + AI 落地)

[ ] Day 35: 简历最终打磨与投递准备