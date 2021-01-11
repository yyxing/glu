package cache

import "time"

type Cache interface {
	Put(key string, value Value) bool
	Get(key string) (interface{}, bool)
	Del(key string) bool
	Len() int
	Keys() []string
	Size() uint64
}

// 存储到队列和表中的具体数据结构
type entry struct {
	key        string
	value      Value
	visitCount uint
}

// 存储到队列和表中的具体数据结构
type redisEntry struct {
	key   string
	value Value
	idle  time.Time
}

// TODO 所有的value必须实现该接口 删除限制 改用自己获取
type Value interface {
	Len() int
}
