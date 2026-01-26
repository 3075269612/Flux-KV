# 🚀 35天 Golang 后端 + AI 全栈突击计划

## 🎯 核心目标
1. **项目**: 手写分布式 KV 存储 + 高性能网关 + AI RAG 知识库
2. **算法**: LeetCode Hot 100 (每日 3 题)
3. **基础**: 计算机网络 / 操作系统 / MySQL / Redis 原理

## 📅 每日作息 (Hardcore Mode)
- **08:00 - 09:00**: 早读 (CS 基础/八股文)
- **09:00 - 10:30**: 算法 (Hot 100 x 3，必须手撕)
- **10:30 - 12:30**: 核心开发 (Golang)
- **14:00 - 18:00**: 调试/Linux/中间件集成
- **20:00 - 22:00**: 复盘 & 笔记整理

---

## 🗺️ 阶段一：分布式存储引擎 (Day 1-10)
- [x] **Day 1**: ~~线程安全 Map~~ -> 工程基础设施搭建 (Layout/Viper/Zap/Gin) (已完成 ✅)

- [x] **Day 1**: 线程安全 Map (sync.RWMutex) + 竞态检测 test (已达成 ✅)
- [x] **Day 2**: 实现 TTL 过期清理机制 (Lazy + Active GC + Double Check 锁优化) ✅
- [x] **Day 3**: AOF 持久化 (文件 IO + 启动恢复) (已完成 ✅)
- [x] **Day 4**: TCP 服务端搭建 (自定义协议解决粘包/拆包) (已完成 ✅)
- [x] **Day 5**: 客户端 SDK 封装 (连接池基础/长连接/集成测试) (已完成 ✅)
- [x] **Day 6**: gRPC 改造 (Protobuf 定义) (已完成 ✅)
- [x] **Day 7**: gRPC 双向通信调试 & 交互式 CLI 开发 (已完成 ✅)
- [x] **Day 8**: 一致性哈希算法 (核心) (虚拟节点 + 动态扩缩容) (已完成 ✅)
- [ ] **Day 9**: 分布式节点注册
- [ ] **Day 10**: 阶段总结 & Benchmark 压测

## 🗺️ 阶段二：高性能 API 网关 (Day 11-20)
- [ ] **Day 11**: HTTP Server 搭建 (Gin/net/http)
- [ ] **Day 12**: 反向代理逻辑 (Proxy to gRPC)
- [ ] **Day 13**: 中间件架构 (Logger)
- [ ] **Day 14**: 异常恢复 (Recovery)
- [ ] **Day 15**: 全局限流 (Token Bucket)
- [ ] **Day 16**: 熔断器 (Hystrix 模式)
- [ ] **Day 17**: 负载均衡策略
- [ ] **Day 18**: SingleFlight 防击穿
- [ ] **Day 19**: 性能分析 (Pprof) & 优化
- [ ] **Day 20**: 阶段复盘 & 压力测试 (wrk/ab)

## 🗺️ 阶段三：AI RAG 赋能 (Day 21-30)
- [ ] **Day 21**: DeepSeek/OpenAI API 接入
- [ ] **Day 22**: 流式响应 (Stream) 处理
- [ ] **Day 23**: 文本向量化 (Embeddings)
- [ ] **Day 24**: 向量相似度计算
- [ ] **Day 25**: RAG 流程跑通
- [ ] **Day 26**: AI 结果缓存 (KV Store 集成)
- [ ] **Day 27**: Docker 容器化 (Dockerfile)
- [ ] **Day 28**: Docker Compose 编排
- [ ] **Day 29**: 系统联调
- [ ] **Day 30**: 最终测试

## 📝 阶段四：简历与面试 (Day 31-35)
- [ ] **Day 31**: 架构图绘制 (Excalidraw)
- [ ] **Day 32**: 难点/坑点文档整理
- [ ] **Day 33**: 模拟面试 (基础篇)
- [ ] **Day 34**: 模拟面试 (项目篇)
- [ ] **Day 35**: 简历最终打磨