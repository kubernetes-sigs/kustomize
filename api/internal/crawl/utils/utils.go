package utils

import "sync"

type SeenMap struct {
	data map[string]string
	lock sync.RWMutex
}

// TODO: add lock to avoid race condition
func (seen SeenMap) Seen(item string) bool {
	seen.lock.RLock()
	_, ok := seen.data[item]
	seen.lock.RUnlock()
	return ok
}

func (seen SeenMap) Set(k, v string) {
	seen.lock.Lock()
	seen.data[k] = v
	seen.lock.Unlock()
}

// The caller should make sure that key is in the map.
func (seen SeenMap) Value(k string) string {
	seen.lock.RLock()
	v := seen.data[k]
	seen.lock.RUnlock()
	return v
}

func NewSeenMap() SeenMap {
	return SeenMap{
		data: make(map[string]string),
		lock: sync.RWMutex{},
	}
}
