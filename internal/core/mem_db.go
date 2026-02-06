package core

import (
	"Flux-KV/internal/aof"
	"Flux-KV/internal/config"
	"log"
	"sync"
	"time"
)

// 定义分片数量，在大并发下足够减少锁冲突
const ShardCount = 256

// Item 封装了值和过期时间
type Item struct {
	Val      any
	ExpireAt int64
}

// 定义分片结构
type shard struct {
	mu   sync.RWMutex
	data map[string]*Item
}

// MemDB 内存数据库核心结构
type MemDB struct {
	shards []*shard

	aofHandler *aof.AofHandler // 新增：持有AOF操作对象
}

// 实现 FNV-1a 哈希算法
// 公式：hash = (hash ^ byte) * prime
func fnv32(key string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	hash := uint32(offset32)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= prime32
	}
	return hash
}

// getShard 根据 Key 路由到指定的分片
func (db *MemDB) getShard(key string) *shard {
	hash := fnv32(key)
	return db.shards[hash%ShardCount]
}

func NewMemDB(cfg *config.Config) *MemDB {
	db := &MemDB{
		shards: make([]*shard, ShardCount),
	}

	// 初始化所有分片
	for i := 0; i < ShardCount; i++ {
		db.shards[i] = &shard{
			data: make(map[string]*Item),
		}
	}

	// 初始化 AOF 模块
	if cfg.AOF.Filename != "" {
		handler, err := aof.NewAofHandler(cfg.AOF.Filename)
		if err != nil {
			log.Fatal(err)
		}
		db.aofHandler = handler

		// 启动时立刻恢复数据
		db.loadFromAof()
	}

	return db
}

// loadFromAof 从 AOF 文件恢复数据
func (db *MemDB) loadFromAof() {
	if db.aofHandler == nil {
		return
	}

	// 读取所有命令
	cmds, err := db.aofHandler.ReadAll()
	if err != nil {
		log.Printf("Read AOF failure: %v", err)
		return
	}

	// 重放命令，针对每个 Key 找分片锁
	for _, cmd := range cmds {
		s := db.getShard(cmd.Key)
		s.mu.Lock()
		switch cmd.Type {
		case "set":
			s.data[cmd.Key] = &Item{
				Val:      cmd.Value,
				ExpireAt: 0,
			}
		case "del":
			delete(s.data, cmd.Key)
		}
		s.mu.Unlock()
	}
}

// Set 写入数据，支持过期时间(ttl: time to live)
// ttl = 0 表示永不过期
func (db *MemDB) Set(key string, val any, ttl time.Duration) {
	// 1. 定位分片
	s := db.getShard(key)

	var expireAt int64 = 0
	if ttl > 0 {
		expireAt = time.Now().Add(ttl).UnixNano()
	}

	// 2. 分片加锁（细粒度）
	s.mu.Lock()
	s.data[key] = &Item{val, expireAt}
	s.mu.Unlock()

	// 3. 写 AOF
	if db.aofHandler != nil {
		cmd := aof.Cmd{
			Type:  "set",
			Key:   key,
			Value: val,
		}
		_ = db.aofHandler.Write(cmd)
	}
}

// Get 获取数据（实现惰性删除）
func (db *MemDB) Get(key string) (any, bool) {
	s := db.getShard(key)

	// 1. 分片读锁
	s.mu.RLock()
	item, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return nil, false
	}

	// 2. 惰性删除判断
	if item.ExpireAt > 0 && time.Now().UnixNano() > item.ExpireAt {
		// 发现过期，惰性删除
		s.mu.Lock()
		defer s.mu.Unlock()

		// Double Check双重检查，防止加锁间隙被其他协程处理
		newItem, exists := s.data[key]
		if !exists {
			// 已经被别人删了
			return nil, false
		}

		// 依然存在，且依然是过期状态，真删
		if newItem.ExpireAt > 0 && time.Now().UnixNano() > newItem.ExpireAt {
			delete(s.data, key)
			return nil, false
		}

		// 第一次看过期，第二次看续命
		return newItem.Val, true
	}

	return item.Val, true
}

// Del 手动删除数据
func (db *MemDB) Del(key string) {
	s := db.getShard(key)

	s.mu.Lock()
	defer s.mu.Unlock()

	// 删内存
	delete(s.data, key)

	// 写 AOF
	if db.aofHandler != nil {
		cmd := aof.Cmd{
			Type: "del",
			Key:  key,
		}
		_ = db.aofHandler.Write(cmd)
	}
}

// 关闭数据库
func (db *MemDB) Close() {
	if db.aofHandler != nil {
		_ = db.aofHandler.Close()
	}
}

// StartGC 启动定期清理（Garbage Collection）
// interval: 清理间隔，例如1秒
func (db *MemDB) StartGC(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			db.activeCleanup()
		}
	}()
}

// activeCleanup 遍历 map 清理过期数据
func (db *MemDB) activeCleanup() {
	now := time.Now().UnixNano()

	// 遍历每一个分片
	for _, s := range db.shards {
		// 1. 快速读锁检查
		s.mu.RLock()
		var expireKeys []string
		for key, item := range s.data {
			if item.ExpireAt > 0 && now > item.ExpireAt {
				expireKeys = append(expireKeys, key)
			}
		}
		s.mu.RUnlock()

		// 2. 如果有需要删除的 Key，再加写锁
		if len(expireKeys) > 0 {
			s.mu.Lock()
			defer s.mu.Unlock()

			for _, key := range expireKeys {
				// Double Check
				item, exists := s.data[key]
				if exists && item.ExpireAt > 0 && time.Now().UnixNano() > item.ExpireAt {
					delete(s.data, key)
				}
			}
		}
	}
}
