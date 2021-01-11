package cache

import (
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"
)

const (
	EVPOOLSize = 16
)

var (
	seed = rand.NewSource(time.Now().UnixNano())
)

type eliminatedKeyPool struct {
	entries []*redisEntry
	size    int
}

func (p *eliminatedKeyPool) Len() int {
	return p.size
}
func (p *eliminatedKeyPool) Swap(i, j int) {
	entries := p.entries
	entries[i], entries[j] = entries[j], entries[i]
}
func (p *eliminatedKeyPool) Less(i, j int) bool {
	entries := p.entries
	now := time.Now()
	return now.Sub(entries[i].idle).Nanoseconds() < now.Sub(entries[j].idle).Nanoseconds()
}

type RedisLRUCache struct {
	// 缓存总大小
	maxBytes uint64
	// 缓存使用大小
	useBytes uint64
	// 缓存映射
	cache map[string]*redisEntry
	// 清理缓存历史定时器
	ticker time.Ticker
	// 读写锁
	mux sync.RWMutex
	// 节点池 节省内存
	ePool sync.Pool
	// key pool
	keyPool *eliminatedKeyPool
	// 取样大小 默认为5
	maxSamples int
	// size
	size int
	// 删除触发器
	OnEvicted func(key string, value Value)
}

func NewRedisLRUCache(capacity uint64, maxSamples int, onEvicted func(string, Value)) Cache {
	return &RedisLRUCache{
		maxBytes: capacity,
		cache:    make(map[string]*redisEntry),
		ePool: sync.Pool{New: func() interface{} {
			return new(redisEntry)
		}},
		keyPool: &eliminatedKeyPool{
			entries: make([]*redisEntry, EVPOOLSize),
			size:    0,
		},
		maxSamples: maxSamples,
		OnEvicted:  onEvicted,
	}
}
func (cache *RedisLRUCache) Put(key string, value Value) bool {
	// 加锁
	cache.mux.Lock()
	defer cache.mux.Unlock()
	if e, ok := cache.cache[key]; ok {
		cache.useBytes += uint64(value.Len() - e.value.Len())
		e.value = value
		e.idle = time.Now()
	} else {
		useBytes := uint64(len(key) + value.Len())
		if useBytes > cache.maxBytes {
			log.Println("Out of Memory Error")
			return false
		}
		// 执行redis lru删除逻辑
		for cache.useBytes+useBytes > cache.maxBytes {
			cache.removeOldest()
		}
		newEntry := cache.ePool.Get().(*redisEntry)
		newEntry.key = key
		newEntry.value = value
		newEntry.idle = time.Now()
		cache.cache[key] = newEntry
		cache.useBytes += useBytes
		cache.size++
	}
	return true
}
func (cache *RedisLRUCache) Get(key string) (interface{}, bool) {
	// 读锁
	cache.mux.RLock()
	defer cache.mux.RUnlock()
	if e, ok := cache.cache[key]; ok {
		e.idle = time.Now()
		return e.value, true
	}
	return nil, false
}
func (cache *RedisLRUCache) Keys() []string {
	keys := make([]string, 0)
	for key := range cache.cache {
		keys = append(keys, key)
	}
	return keys
}
func (cache *RedisLRUCache) samples() []*redisEntry {
	keys := cache.Keys()
	count := cache.maxSamples
	if cache.Len() < cache.maxSamples {
		count = cache.Len()
	}
	samples := make([]*redisEntry, 0)
	bit := make([]byte, cache.Len())
	r := rand.New(seed)
	for count > 0 {
		num := r.Intn(cache.Len())
		if bit[num] == 1 {
			continue
		}
		bit[num] = 1
		count--
		samples = append(samples, cache.cache[keys[num]])
	}
	return samples
}
func (cache *RedisLRUCache) removeOldest() {
	cache.evictionPoolPopulate()
	for i := cache.keyPool.size - 1; i >= 0; i-- {
		cache.delEntry(cache.keyPool.entries[i])
		cache.keyPool.size--
		cache.keyPool.entries[cache.keyPool.size] = nil
	}
}
func (cache *RedisLRUCache) Size() uint64 {
	return cache.useBytes
}
func (cache *RedisLRUCache) evictionPoolPopulate() {
	samples := cache.samples()
	now := time.Now()
	for i := 0; i < len(samples); i++ {
		de := samples[i]
		idle := now.Sub(de.idle).Nanoseconds()
		k := 0
		for k < EVPOOLSize && cache.keyPool.entries[k] != nil &&
			now.Sub(cache.keyPool.entries[k].idle).Nanoseconds() < idle {
			k++
		}
		// 说明没有比当前节点空闲时间小的 表示这次取样是最近访问的
		if k == 0 && cache.keyPool.entries[EVPOOLSize-1] != nil {
			continue
		} else if k < EVPOOLSize && cache.keyPool.size < EVPOOLSize {
			// 有空位可以放入pool中
		} else {
			// 无空位 释放池中内存 释放完后可将当前抽样节点放入最后
			if cache.keyPool.size >= EVPOOLSize {
				delEntry := cache.keyPool.entries[cache.keyPool.size-1]
				cache.delEntry(delEntry)
				cache.keyPool.size--
				cache.keyPool.entries[cache.keyPool.size] = nil
			}
			// 表示当前取样为可以放入池中的 且不为最后
			// 原版c语言利用memmove插入至数组中 并且将k以后的内存向后移动 移出k位置的内存
		}
		// 添加数据至空位处
		cache.keyPool.entries[cache.keyPool.size] = de
		cache.keyPool.size++
		sort.Sort(cache.keyPool)
	}
}

func (cache *RedisLRUCache) delEntry(delEntry *redisEntry) {
	delete(cache.cache, delEntry.key)
	cache.size--
	cache.useBytes -= uint64(len(delEntry.key) + delEntry.value.Len())
	if cache.OnEvicted != nil {
		cache.OnEvicted(delEntry.key, delEntry.value)
	}
	// 将释放的元素加入到pool中
	cache.ePool.Put(delEntry)
}
func (cache *RedisLRUCache) Del(key string) bool {
	// 写锁
	cache.mux.Lock()
	defer cache.mux.Unlock()
	if e, ok := cache.cache[key]; ok {
		cache.delEntry(e)
		return true
	}
	return false
}
func (cache *RedisLRUCache) Len() int {
	return cache.size
}
