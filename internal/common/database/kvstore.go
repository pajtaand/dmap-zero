package database

import (
	"sync"
)

type KVStore struct {
	mu    sync.RWMutex
	store map[string]interface{}
}

func NewKVStore() *KVStore {
	return &KVStore{
		store: map[string]interface{}{},
	}
}

func (kv *KVStore) Set(key string, value interface{}) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.store[key] = value
}

func (kv *KVStore) Get(key string) (interface{}, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	value, ok := kv.store[key]
	return value, ok
}

func (kv *KVStore) Delete(key string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.store, key)
}

func (kv *KVStore) Keys() []string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	keys := make([]string, 0, len(kv.store))
	for k := range kv.store {
		keys = append(keys, k)
	}
	return keys
}

func (kv *KVStore) Count() int {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	return len(kv.store)
}

func (kv *KVStore) Clear() {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.store = map[string]interface{}{}
}
