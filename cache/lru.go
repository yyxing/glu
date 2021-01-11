package cache

import (
	"container/list"
	"sync"
	"time"
)

// lru cache
type LRUCache struct {
	// 缓存总大小
	maxBytes uint64
	// 缓存使用大小
	useBytes uint64
	// 缓存双向队列
	cacheList *list.List
	// 缓存映射
	cache map[string]*list.Element
	// 清理缓存历史定时器
	ticker time.Ticker
	// 读写锁
	mux sync.RWMutex
	// 节点池 节省内存
	ePool sync.Pool
	// 删除触发器
	OnEvicted func(key string, value Value)
}

func NewLRUCache(capacity uint64, onEvicted func(key string, value Value)) Cache {
	return &LRUCache{
		maxBytes:  capacity,
		useBytes:  0,
		cacheList: list.New(),
		cache:     make(map[string]*list.Element),
		ePool: sync.Pool{New: func() interface{} {
			return new(entry)
		}},
		OnEvicted: onEvicted,
	}
}
func (cache *LRUCache) Put(key string, value Value) bool {
	// 加锁
	cache.mux.Lock()
	defer cache.mux.Unlock()
	// 缓存存在 将缓存移到队尾
	if element, ok := cache.cache[key]; ok {
		e := element.Value.(*entry)
		// 将缓存移到队尾
		cache.cacheList.MoveToFront(element)
		// 修改使用大小
		cache.useBytes += uint64(value.Len() - e.value.Len())
		// 值覆盖
		e.value = value
	} else {
		// 新节点添加
		newEntry := cache.ePool.Get().(*entry)
		// 节点赋值
		newEntry.key = key
		newEntry.value = value
		// 插入节点到队列中 并且将映射关系放入map中
		newElement := cache.cacheList.PushFront(newEntry)
		cache.cache[key] = newElement
		// 添加使用内存大小
		cache.useBytes += uint64(len(key) + value.Len())
	}
	// 内存超限 根据lru淘汰数据
	for cache.useBytes > cache.maxBytes {
		cache.removeOldest()
	}
	return true
}
func (cache *LRUCache) Size() uint64 {
	return cache.useBytes
}
func (cache *LRUCache) Get(key string) (interface{}, bool) {
	// 读锁
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
func (cache *LRUCache) Keys() []string {
	keys := make([]string, len(cache.cache))
	for key := range cache.cache {
		if key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}
func (cache *LRUCache) Del(key string) bool {
	// 写锁
	cache.mux.Lock()
	defer cache.mux.Unlock()
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
		// 将释放的元素加入到pool中
		cache.ePool.Put(e)
		return true
	}
	return false
}

// 淘汰最少使用元素
func (cache *LRUCache) removeOldest() {
	// 获取队头元素
	delElement := cache.cacheList.Back()
	if delElement != nil {
		e := delElement.Value.(*entry)
		// 从链表中删除
		cache.cacheList.Remove(delElement)
		// 从map映射删除
		delete(cache.cache, e.key)
		// 减少内存使用
		cache.useBytes -= uint64(e.value.Len() + len(e.key))
		// 调用删除触发器
		if cache.OnEvicted != nil {
			cache.OnEvicted(e.key, e.value)
		}
		cache.ePool.Put(e)
	}
}
func (cache *LRUCache) Len() int {
	return cache.cacheList.Len()
}
