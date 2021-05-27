package cache

import (
	"time"
)

type Memory struct {
	M map[string]interface{}
}

func NewMemory(expire, purge time.Duration) *Memory {
	return &Memory{make(map[string]interface{})}
}

func (m *Memory) Set(key string, val interface{}) error {
	m.M[key] = val
	return nil
}

func (m *Memory) Get(key string) (interface{}, bool) {
	val, ok := m.M[key]
	return val, ok
}

func (m *Memory) Drop(key string) error {
	delete(m.M, key)
	return nil
}
