package core

import (
	"sync"
)

type MemDB struct {
	sync.RWMutex	// 嵌入读写锁，保护下面的map不被乱改

	m map[string]string
}

// 初始化
func NewMemDB() *MemDB {
	return &MemDB{
		m: make(map[string]string),
	}
}

// Set 存数据（写操作）
func (db *MemDB) Set(key string, value string) {
	db.Lock()
	defer db.Unlock()
	db.m[key] = value
}

// Get 取数据（读操作）
func (db *MemDB) Get(key string) (string, bool) {
	db.RLock()
	defer db.RUnlock()
	val, ok := db.m[key]
	return val, ok
}

// Delete 删数据（写操作）
func (db *MemDB) Delete(key string) {
	db.Lock()
	defer db.Unlock()
	delete(db.m, key)
}