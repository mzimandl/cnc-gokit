// Copyright 2022 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2022 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collections

import (
	"encoding/json"
	"sync"
)

type ConcurrentMap[K comparable, T any] struct {
	sync.RWMutex
	data map[K]T
}

func (cm *ConcurrentMap[K, T]) Get(k K) T {
	cm.RLock()
	defer cm.RUnlock()
	return cm.data[k]
}

func (cm *ConcurrentMap[K, T]) GetWithTest(k K) (T, bool) {
	cm.RLock()
	defer cm.RUnlock()
	v, ok := cm.data[k]
	return v, ok
}

func (cm *ConcurrentMap[K, T]) HasKey(k K) bool {
	cm.RLock()
	defer cm.RUnlock()
	_, ok := cm.data[k]
	return ok
}

func (cm *ConcurrentMap[K, T]) Set(k K, v T) {
	cm.Lock()
	defer cm.Unlock()
	cm.data[k] = v
}

func (cm *ConcurrentMap[K, T]) Delete(k K) {
	cm.Lock()
	defer cm.Unlock()
	delete(cm.data, k)
}

// ForEach iterates through all the items. Due to concurrent
// nature - to prevent possible issues and or deadlocks, the iteration
// copies all the keys available in time the method was called and during
// each `yield` call, it tries the obtain a corresponding
// value - which may or may not be available.
// The method acquires locks only for necessary operations inside so
// it should be deadlock-resistant.
func (cm *ConcurrentMap[K, T]) ForEach(yield func(k K, v T, ok bool)) {
	var keys []K
	cm.RLock()
	keys = make([]K, len(cm.data))
	var i int
	for k := range cm.data {
		keys[i] = k
		i++
	}
	cm.RUnlock()
	for _, k := range keys {
		cm.RLock()
		v, ok := cm.data[k]
		cm.RUnlock()
		yield(k, v, ok)
	}
}

func (cm *ConcurrentMap[K, T]) Update(fn func(k K, v T) T) {
	cm.Lock()
	defer cm.Unlock()
	for k, v := range cm.data {
		cm.data[k] = fn(k, v)
	}
}

func (cm *ConcurrentMap[K, T]) Keys() []K {
	cm.RLock()
	defer cm.RUnlock()
	ans := make([]K, len(cm.data))
	i := 0
	for k, _ := range cm.data {
		ans[i] = k
		i++
	}
	return ans
}

func (cm *ConcurrentMap[K, T]) Values() []T {
	cm.RLock()
	defer cm.RUnlock()
	ans := make([]T, len(cm.data))
	i := 0
	for _, v := range cm.data {
		ans[i] = v
		i++
	}
	return ans
}

// AsMap creates a shallow copy of a map wrapped
// by this ConcurrentMap
func (cm *ConcurrentMap[K, T]) AsMap() map[K]T {
	cm.RLock()
	defer cm.RUnlock()
	ans := make(map[K]T)
	for k, v := range cm.data {
		ans[k] = v
	}
	return ans
}

// Len returns number of key-value pairs stored in the map
func (cm *ConcurrentMap[K, T]) Len() int {
	cm.RLock()
	defer cm.RUnlock()
	return len(cm.data)
}

func (cm *ConcurrentMap[K, T]) Filter(fn func(k K, v T) bool) *ConcurrentMap[K, T] {
	ans := make(map[K]T)
	cm.RLock()
	for kx, vx := range cm.data {
		if fn(kx, vx) {
			ans[kx] = vx
		}
	}
	cm.RUnlock()
	return NewConcurrentMapFrom(ans)
}

func (cm *ConcurrentMap[K, T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(cm.AsMap())
}

func NewConcurrentMap[K comparable, T any]() *ConcurrentMap[K, T] {
	return &ConcurrentMap[K, T]{
		data: make(map[K]T),
	}
}

func NewConcurrentMapFrom[K comparable, T any](data map[K]T) *ConcurrentMap[K, T] {
	return &ConcurrentMap[K, T]{
		data: data,
	}
}

func NewConcurrentMapFromJSON[K comparable, T any](data []byte) (*ConcurrentMap[K, T], error) {
	data2 := make(map[K]T)
	err := json.Unmarshal(data, &data2)
	if err != nil {
		return nil, err
	}
	return &ConcurrentMap[K, T]{
		data: data2,
	}, nil
}
