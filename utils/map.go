package utils

import (
	"sync"
	"sync/atomic"
)

// AdvanceMap wrap arround sync map in go
type AdvanceMap struct {
	count int64     // length of the map
	items *sync.Map // store all data
	mutex sync.RWMutex
}

// NewAdvanceMap linter
func NewAdvanceMap() *AdvanceMap {
	return &AdvanceMap{
		count: 0,
		items: &sync.Map{},
	}
}

func (a *AdvanceMap) getCount() *int64 {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return &a.count
}

func (a *AdvanceMap) decreCount() {
	atomic.AddInt64(a.getCount(), -int64(1))
}

func (a *AdvanceMap) increCount() {
	atomic.AddInt64(a.getCount(), 1)
}

func (a *AdvanceMap) getItems() *sync.Map {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	return a.items
}

// Len return len of sync map
func (a *AdvanceMap) Len() int {
	return int(a.count)
}

// Set set string key with interface value
func (a *AdvanceMap) Set(key string, value interface{}) {
	if items := a.getItems(); items != nil {
		items.Store(key, value)
		a.increCount()
	}
}

// Get get item with key string
func (a *AdvanceMap) Get(key string) (interface{}, bool) {
	if items := a.getItems(); items != nil {
		return items.Load(key)
	}
	return nil, false
}

// Delete map with key string
func (a *AdvanceMap) Delete(key interface{}) {
	if items := a.getItems(); items != nil {
		items.Delete(key)
		a.decreCount()
	}
}

// GetKeys get all keys of map
func (a *AdvanceMap) GetKeys() []string {
	temp := make([]string, 0)
	a.Iter(func(key, value interface{}) bool {
		item, ok := key.(string)
		if ok {
			temp = append(temp, item)
		}
		return true
	})
	return temp
}

// Iter go though a map
func (a *AdvanceMap) Iter(callBack func(key, value interface{}) bool) {
	if items := a.getItems(); items != nil {
		items.Range(callBack)
	}
}

// Capture a current map
// Warning: Current data will be delete after using it
func (a *AdvanceMap) Capture() map[string]interface{} {
	tempMap := make(map[string]interface{}, 0)
	tempFunc := func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if ok {
			tempMap[keyStr] = value
		}
		return true
	}

	if items := a.getItems(); items != nil {
		items.Range(func(key, value interface{}) bool {
			// add key value
			tempFunc(key, value)
			// delete value
			a.Delete(key)
			return true
		})
	}
	return tempMap
}

// ToMap returning map to this value
func (a *AdvanceMap) ToMap() map[string]interface{} {
	tempMap := make(map[string]interface{}, 0)
	if items := a.getItems(); items != nil {
		items.Range(func(key, value interface{}) bool {
			keyStr, ok := key.(string)
			if ok {
				tempMap[keyStr] = value
			}
			return true
		})
	}
	return tempMap
}
