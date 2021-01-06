package cache

import (
	"container/list"
	"sync"
	"time"
)

type LRUKCache struct {
	// 缓存总大小
	maxBytes uint64
	// 缓存使用大小
	useBytes uint64
	// 缓存总大小
	historyMaxBytes uint64
	// 缓存使用大小
	historyUseBytes uint64
	// 缓存双向队列
	cacheList *list.List
	// 缓存映射
	cache map[string]*list.Element
	// 缓存历史表
	history map[string]*list.Element
	// 缓存历史队列
	historyList *list.List
	// 清理缓存历史定时器
	ticker time.Ticker
	// lru-k的K大小
	k uint
	// 删除触发器
	OnEvicted func(key string, value Value)
	// 锁
	mux sync.RWMutex
}

// 添加更新数据
func (cache *LRUKCache) Put(key string, value Value) bool {
	cache.mux.Lock()
	defer cache.mux.Unlock()
	if cache.cache == nil || cache.cacheList == nil ||
		cache.history == nil || cache.historyList == nil {
		cache.lazyInit()
	}
	// 若节点存在则将节点的值进行 更新同时将元素移到队尾
	if element, ok := cache.cache[key]; ok {
		kv := element.Value.(*entry)
		cache.useBytes += uint64(value.Len() - kv.value.Len())
		kv.value = value
		cache.cacheList.MoveToFront(element)
	} else if element, ok := cache.history[key]; ok {
		kv := element.Value.(*entry)
		kv.visitCount++
		kv.value = value
		// 访问次数足够 加进缓存页面中
		if kv.visitCount >= cache.k {
			// 从历史队列中删除
			delResult := cache.delHistory(key)
			if !delResult {
				return false
			}
			cache.putCache(key, value)
		} else {
			cache.historyList.MoveToFront(element)
		}
	} else {
		// 表示新增缓存 加到缓存页面中
		cache.putHistory(key, value)
	}
	return true
}

// 获取数据
func (cache *LRUKCache) Get(key string) (interface{}, bool) {
	cache.mux.RLock()
	defer cache.mux.RUnlock()
	if element, ok := cache.cache[key]; ok {
		e := element.Value.(*entry)
		// 若缓存存在则将命中的缓存移到队尾
		cache.cacheList.MoveToFront(element)
		return e.value, true
	}
	return nil, false
}

// 删除数据
func (cache *LRUKCache) Del(key string) bool {
	cache.mux.Lock()
	defer cache.mux.Unlock()
	return cache.delCache(key) && cache.delHistory(key)
}

// 根据策略删除缓存
func (cache *LRUKCache) removeOldest() {
	delElement := cache.cacheList.Back()
	if delElement != nil {
		e := delElement.Value.(*entry)
		delete(cache.cache, e.key)
		cache.cacheList.Remove(delElement)
		cache.useBytes -= uint64(len(e.key) + e.value.Len())
		// 调用删除触发器
		if cache.OnEvicted != nil {
			cache.OnEvicted(e.key, e.value)
		}
	}
}

// 根据策略删除缓存
func (cache *LRUKCache) removeHistoryOldest() {
	delElement := cache.historyList.Back()
	if delElement != nil {
		e := delElement.Value.(*entry)
		delete(cache.history, e.key)
		cache.historyList.Remove(delElement)
		cache.historyUseBytes -= uint64(len(e.key) + e.value.Len())
	}
}

// 添加进历史页中
func (cache *LRUKCache) putHistory(key string, value Value) {
	element := cache.historyList.PushFront(&entry{key: key, value: value, visitCount: 1,
		lastVisitTime: time.Now()})
	cache.history[key] = element
	cache.historyUseBytes += uint64(len(key) + value.Len())
	for cache.historyUseBytes > cache.historyMaxBytes {
		cache.removeHistoryOldest()
	}
}

// 添加进缓存页中
func (cache *LRUKCache) putCache(key string, value Value) {
	element := cache.cacheList.PushFront(&entry{key: key, value: value})
	cache.cache[key] = element
	cache.useBytes += uint64(len(key) + value.Len())
	// 大小超容 执行lru-k逻辑
	for cache.useBytes > cache.maxBytes {
		cache.removeOldest()
	}
}

// 删除历史页中数据
func (cache *LRUKCache) delHistory(key string) bool {
	if ele, ok := cache.history[key]; ok {
		e := ele.Value.(*entry)
		delete(cache.history, key)
		cache.historyList.Remove(ele)
		cache.historyUseBytes -= uint64(len(e.key) + e.value.Len())
		return true
	}
	return false
}
func (cache *LRUKCache) Keys() []string {
	keys := make([]string, len(cache.cache))
	for key := range cache.cache {
		if key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}

// 删除缓存页面
func (cache *LRUKCache) delCache(key string) bool {
	if element, ok := cache.cache[key]; ok {
		e := element.Value.(*entry)
		// 删除map映射
		delete(cache.cache, key)
		// 删除链表节点
		cache.cacheList.Remove(element)
		// 内存使用更改
		cache.useBytes -= uint64(e.value.Len() + len(e.key))
		// 调用删除触发器
		if cache.OnEvicted != nil {
			cache.OnEvicted(e.key, e.value)
		}
		return true
	}
	return false
}

// 获取缓存大小
func (cache *LRUKCache) Len() int {
	return cache.cacheList.Len()
}

// 延时加载
func (cache *LRUKCache) lazyInit() {
	if cache.historyList == nil {
		cache.historyList = list.New()
	}
	if cache.cacheList == nil {
		cache.cacheList = list.New()
	}
	if cache.cache == nil {
		cache.cache = make(map[string]*list.Element)
	}
	if cache.history == nil {
		cache.history = make(map[string]*list.Element)
	}
}

func NewLRUKCache(k uint, capacity uint64, OnEvicted func(key string, value Value)) Cache {

	return &LRUKCache{
		k:               k,
		maxBytes:        capacity,
		useBytes:        0,
		historyUseBytes: 0,
		historyMaxBytes: capacity,
		OnEvicted:       OnEvicted,
	}
}
