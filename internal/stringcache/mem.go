package stringcache

import (
	"time"
)

type Memory struct {
	M map[string]string
}

func NewMemory(expire, purge time.Duration) *Memory {
	return &Memory{make(map[string]string)}
}

func (m *Memory) Set(key, val string) error {
	m.M[key] = val
	return nil
}

func (m *Memory) Get(key string) (string, bool) {
	val, ok := m.M[key]
	return val, ok
}
