package utils

import (
	"sync"

	"github.com/thoas/go-funk"
)

type SMap struct {
	m   map[interface{}]interface{}
	mtx sync.RWMutex
}

func NewMap() *SMap {
	return &SMap{m: make(map[interface{}]interface{})}
}

func (m *SMap) Clear() {
	m.m = make(map[interface{}]interface{})
}

func (m *SMap) Add(key, val interface{}) {
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.m[key] = val
}

func (m *SMap) Delete(key interface{}) {
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.m, key)
}

func (m *SMap) Get(key interface{}) interface{} {
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()

	val, ok := m.m[key]
	if !ok {
		return nil
	}

	return val
}

func (m *SMap) Keys() []interface{} {
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()

	return funk.Keys(m.m).([]interface{})
}

func (m *SMap) Values() []interface{} {
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()

	return funk.Values(m.m).([]interface{})
}

func (m *SMap) PopValues() []interface{} {
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}

	m.mtx.RLock()
	defer m.mtx.RUnlock()

	vals := funk.Values(m.m).([]interface{})
	m.Clear()

	return vals
}
