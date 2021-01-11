package cache

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}
func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(30, func(key string, value Value) {
		log.Printf("%s is deleted", key)
	})
	cache.Put("key1", String("value1"))
	cache.Put("key2", String("value2"))
	cache.Put("key3", String("value3"))
	value, isSuccess := cache.Get("key1")
	log.Println(value, isSuccess)
	value, isSuccess = cache.Get("key2")
	log.Println(value, isSuccess)
	value, isSuccess = cache.Get("key3")
	log.Println(value, isSuccess)
	cache.Put("key4", String("value4"))
	value, isSuccess = cache.Get("key1")
	log.Println(value, isSuccess)
	value, isSuccess = cache.Get("key2")
	log.Println(value, isSuccess)
	value, isSuccess = cache.Get("key3")
	log.Println(value, isSuccess)
}
func TestLRUKCache(t *testing.T) {
	cache := NewLRUKCache(2, 30, func(key string, value Value) {
		fmt.Printf("onEvict: k: %v, v: %v\n", key, value)
	})

	cache.Put("key1", String("value1"))
	if _, hit := cache.Get("key1"); hit {
		t.Error("should not get key1 hit")
	}
	cache.Put("key2", String("value2"))
	if _, hit := cache.Get("key2"); hit {
		t.Error("should not get key2 hit")
	}
	cache.Put("key2", String("value2"))
	if _, hit := cache.Get("key2"); !hit {
		t.Error("should get key2 hit")
	}
	cache.Put("key3", String("value3"))
	cache.Put("key3", String("value3"))
	if _, hit := cache.Get("key1"); hit {
		t.Error("could not get key1 hit")
	}
	if _, hit := cache.Get("key3"); !hit {
		t.Error("should get key3 hit")
	}
	// update cache
	cache.Put("key3", String("value33"))
	if v, hit := cache.Get("key3"); !hit || v.(String) != "value33" {
		t.Error("should get key3 hit and value should be value33")
	}
	if _, hit := cache.Get("key2"); !hit {
		t.Error("should get key2 hit")
	}

	// trigger replacing
	cache.Put("key4", String("value4"))
	cache.Put("key4", String("value4"))
	println("cur length: ", cache.Len())
	if _, hit := cache.Get("key4"); !hit {
		t.Error("should get 4 hit")
	}
	if _, hit := cache.Get("key3"); hit {
		t.Error("should not get key3 hit")
	}
	fmt.Println(cache.Keys())
}
func BenchmarkRedisLRUCache(b *testing.B) {
	b.ResetTimer()
	cache := NewRedisLRUCache(1024*1024*1024, 5, nil)
	r := rand.New(seed)
	//fmt.Println(b.N)
	for i := 1; i <= b.N; i++ {
		cache.Put("key"+strconv.FormatInt(int64(r.Intn(b.N)), 10), String("value"+strconv.FormatInt(int64(i), 10)))
	}
}

func BenchmarkLRUCache(b *testing.B) {
	b.ResetTimer()
	cache := NewLRUCache(1024*1024*10, nil)
	r := rand.New(seed)
	for i := 1; i <= b.N; i++ {
		cache.Put("key"+strconv.FormatInt(int64(r.Intn(b.N)), 10), String("value"+strconv.FormatInt(int64(i), 10)))
	}
	//fmt.Println(cache.Len())
	//b.Run("TestGet", func(b *testing.B) {
	//	for i := 1; i <= b.N; i++ {
	//		cache.Get("key" + strconv.FormatInt(int64(r.Intn(i)),10))
	//	}
	//})
}
func BenchmarkLRUKCache(b *testing.B) {
	b.ResetTimer()
	cache := NewLRUKCache(2, 1024*1024*10, nil)
	r := rand.New(seed)
	for i := 1; i <= b.N; i++ {
		cache.Put("key"+strconv.FormatInt(int64(r.Intn(b.N)), 10), String("value"+strconv.FormatInt(int64(i), 10)))
	}
	//fmt.Println(cache.Len())
	//b.Run("TestGet", func(b *testing.B) {
	//	for i := 1; i <= b.N; i++ {
	//		cache.Get("key" + strconv.FormatInt(int64(r.Intn(i)),10))
	//	}
}
