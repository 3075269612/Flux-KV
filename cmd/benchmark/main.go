package main

import (
	"Flux-KV/pkg/client"
	"Flux-KV/pkg/consistent"
	"Flux-KV/pkg/discovery"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

var (
	concurrency = flag.Int("c", 100, "并发数 (Goroutines)")
	totalReq    = flag.Int("n", 100_000, "总请求数")
	endpoints   = flag.String("etcd", "localhost:2379", "Etcd 地址")

	// 统计指标
	successCount int64
	failCount    int64

	clients   = make(map[string]*client.Client)
	ring      *consistent.Map
	clientsMu sync.RWMutex
)

func main() {
	flag.Parse()
	fmt.Printf("🚀 开始压测: %d 并发, 目标 %d 请求, Etcd: %s\n", *concurrency, *totalReq, *endpoints)

	// 1. 初始化客户端
	ring = consistent.New(20, nil)

	d, err := discovery.NewDiscovery([]string{*endpoints})
	if err != nil {
		log.Fatalf("无法连接 Etcd: %v", err)
	}
	// 监听节点上下线（与服务器注册的前缀保持一致）
	err = d.WatchService("/services/kv-service/", func(k, v string) {
		clientsMu.Lock()
		defer clientsMu.Unlock()
		addr := v
		if _, ok := clients[addr]; !ok {
			if cli, err := client.NewDirectClient(addr); err == nil {
				clients[addr] = cli
				ring.Add(addr)
			}
		}
	}, func(k, v string) {
		clientsMu.Lock()
		defer clientsMu.Unlock()
		// v 就是地址，直接使用
		addr := v
		if cli, ok := clients[addr]; ok {
			cli.Close()
			delete(clients, addr)
			ring.Remove(addr)
		}
	})
	if err != nil {
		log.Fatalf("无法监听服务: %v", err)
	}

	time.Sleep(1 * time.Second)

	// 2. 启动监控协程（每秒打印 QPS）
	go monitor()

	// 3. 启动并发 Workers
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(*concurrency)

	// 计算每个 Worker 需要完成的任务量
	reqPerWorker := *totalReq / *concurrency

	for i := 0; i < *concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			runWorker(reqPerWorker, workerID)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// 4. 最终报告
	printReport(duration)
}

// 模拟单个用户的行为
func runWorker(count int, workerID int) {
	// 预先生成随机 Key 前缀，模拟不同数据
	keyPrefix := fmt.Sprintf("user_%d_,", workerID)

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s%d", keyPrefix, i)
		value := fmt.Sprintf("value_%d", rand.Intn(1000))

		// 每次操作前，问一下哈希环：这个 key 归谁管？
		clientsMu.RLock()
		nodeAddr := ring.Get(key)
		targetClient, ok := clients[nodeAddr]
		clientsMu.RUnlock()

		if !ok {
			atomic.AddInt64(&failCount, 1)
			continue
		}

		// 测试 Set
		err := targetClient.Set(key, value)
		if err != nil {
			atomic.AddInt64(&failCount, 1)

			// 👇👇👇 必须把这行注释打开！让我们看到报错信息！👇👇👇
			log.Printf("Set Error: %v", err)
		} else {
			atomic.AddInt64(&successCount, 1)
		}
	}
}

// 监控器：每秒输出当前 QPS
func monitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastCount int64
	for range ticker.C {
		current := atomic.LoadInt64(&successCount)
		diff := current - lastCount
		lastCount = current
		fmt.Printf("🔥 QPS: %d | 成功: %d | 失败: %d\n", diff, current, atomic.LoadInt64(&failCount))
	}
}

func printReport(d time.Duration) {
	total := atomic.LoadInt64(&successCount) + atomic.LoadInt64(&failCount)

	qps := float64(total) / d.Seconds()

	fmt.Println("\n--- 🏁 压测报告 ---")
	fmt.Printf("耗时: %v\n", d)
	fmt.Printf("总请求: %d\n", total)
	fmt.Printf("成功率: %.2f%%\n", float64(successCount)/float64(total)*100)
	fmt.Printf("平均 QPS: %.2f\n", qps)
	fmt.Println("-------------------")
}
