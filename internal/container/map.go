// Package container 实现一些容器
package container

import (
	"sync"
)

// Map 是一个线程安全的泛型 map
type Map[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

// NewMap 创建一个新的带锁 map
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		data: make(map[K]V),
	}
}

// Set 设置键值对
func (m *Map[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Get 获取值，第二个返回值表示是否存在
func (m *Map[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[key]
	return value, exists
}

// Delete 删除键值对
func (m *Map[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// Exists 检查键是否存在
func (m *Map[K, V]) Exists(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.data[key]
	return exists
}

// Len 返回 map 长度
func (m *Map[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

// Clear 清空 map
func (m *Map[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[K]V)
}

// Keys 返回所有键
func (m *Map[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回所有值
func (m *Map[K, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]V, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

// Range 遍历 map（注意：回调中不应修改 map）
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.data {
		if !f(k, v) {
			break
		}
	}
}

// GetOrSet 如果键不存在则设置并返回新值，否则返回已有值
func (m *Map[K, V]) GetOrSet(key K, value V) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, exists := m.data[key]; exists {
		return v, true
	}
	m.data[key] = value
	return value, false
}

// Update 更新已存在的键的值
func (m *Map[K, V]) Update(key K, value V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[key]; exists {
		m.data[key] = value
		return true
	}
	return false
}

// BatchSet 批量设置键值对
func (m *Map[K, V]) BatchSet(items map[K]V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range items {
		m.data[k] = v
	}
}

// BatchDelete 批量删除键
func (m *Map[K, V]) BatchDelete(keys []K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.data, k)
	}
}
