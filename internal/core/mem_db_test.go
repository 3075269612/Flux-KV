package core

import (
	"sync"
	"testing"
)

// TestConcurrentSafety 专门测试并发安全性
// 如果代码写得不对，这个测试跑起来会直接报错或者卡死
func TestConcurrentSafety(t *testing.T) {
	// 1. 创建一个新的数据库实例
	db := NewMemDB()
	
	// WaitGroup 是个计数器，用来等待所有小弟（协程）干完活
	var wg sync.WaitGroup

	// ===========================
	// 模拟 100 个写操作（疯狂写入）
	// ===========================
	for i := 0; i < 100; i++ {
		wg.Add(1) // 计数器 +1
		
		// go func() 启动一个协程（相当于雇个临时工）
		go func(i int) {
			defer wg.Done() // 干完活后，计数器 -1
			db.Set("key", "value") // 每个人都在抢着写！
		}(i)
	}

	// ===========================
	// 模拟 100 个读操作（疯狂读取）
	// ===========================
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db.Get("key") // 每个人都在抢着看！
		}()
	}

	// 等待上面 200 个协程全部结束
	wg.Wait()
	
	// 如果能运行到这里，说明没有崩溃！
	t.Log("✅ 恭喜！你的 Map 在 200 个并发协程的轰炸下安然无恙！")
}