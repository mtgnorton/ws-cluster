package lru

import (
	"errors"

	"sync"
)

// Simple 简单的 LRU 算法
type Simple[K comparable, V any] struct {
	size  int
	items map[K]*Entry[K, V]
	ll    *LinkedList[K, V]
	lock  sync.RWMutex
}

// NewSimple 创建一个新的 Simple LRU
func NewSimple[K comparable, V any](size int) (lru *Simple[K, V], err error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	return &Simple[K, V]{
		size:  size,
		items: make(map[K]*Entry[K, V]),
		ll:    NewLinkedList[K, V](),
	}, nil
}

func (s *Simple[K, V]) Add(key K, value V) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.add(key, value)
}

func (s *Simple[K, V]) Get(key K) (value V, ok bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.get(key)
}

func (s *Simple[K, V]) Contains(key K) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.contains(key)
}

func (s *Simple[K, V]) Peek(key K) (value V, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.peek(key)
}

func (s *Simple[K, V]) Remove(key K) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.remove(key)
}

func (s *Simple[K, V]) RemoveOldest() (key K, value V, isFound bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.removeOldest()
}

func (s *Simple[K, V]) GetNewest() (key K, value V, isFound bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.getNewest()
}

func (s *Simple[K, V]) GetOldest() (key K, value V, isFound bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.getOldest()
}

func (s *Simple[K, V]) Keys() []K {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.keys()
}

func (s *Simple[K, V]) Values() []V {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.values()
}

func (s *Simple[K, V]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.len()
}

func (s *Simple[K, V]) Cap() int {
	return s.cap()
}

func (s *Simple[K, V]) Resize(size int) (evicted int) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.resize(size)
}

func (s *Simple[K, V]) Purge() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.purge()
}

func (s *Simple[K, V]) String() string {
	return "Simple LRU: " + s.ll.String()
}

func (s *Simple[K, V]) add(key K, value V) bool {
	if entry, ok := s.items[key]; ok {
		s.ll.MoveToFront(entry)
		entry.Val = value
		return false
	}
	entry := s.ll.InsertAtHead(key, value)
	s.items[key] = entry
	evict := s.ll.Len() > s.size
	if evict {
		delete(s.items, s.ll.Last().Key)
		s.ll.RemoveAtTail()
	}
	return evict
}
func (s *Simple[K, V]) get(key K) (value V, ok bool) {
	if entry, ok := s.items[key]; ok {
		s.ll.MoveToFront(entry)
		return entry.Val, true
	}
	return
}

func (s *Simple[K, V]) contains(key K) (ok bool) {
	_, ok = s.items[key]
	return
}

func (s *Simple[K, V]) peek(key K) (value V, ok bool) {
	if entry, ok := s.items[key]; ok {
		return entry.Val, true
	}
	return
}
func (s *Simple[K, V]) remove(key K) bool {
	if entry, ok := s.items[key]; ok {
		s.ll.RemoveEntry(entry)
		delete(s.items, key)
		return true
	}
	return false
}
func (s *Simple[K, V]) removeOldest() (key K, value V, isFound bool) {
	if entry := s.ll.Last(); entry != nil {
		s.ll.RemoveEntry(entry)
		delete(s.items, entry.Key)
		return entry.Key, entry.Val, true
	}
	return
}

func (s *Simple[K, V]) getNewest() (key K, value V, isFound bool) {
	if entry := s.ll.First(); entry != nil {
		return entry.Key, entry.Val, true
	}
	return
}
func (s *Simple[K, V]) getOldest() (key K, value V, isFound bool) {
	if entry := s.ll.Last(); entry != nil {
		return entry.Key, entry.Val, true
	}
	return
}

func (s *Simple[K, V]) keys() []K {
	keys := make([]K, s.ll.Len())
	i := 0
	for e := s.ll.Last(); e != nil; e = e.Prev() {
		keys[i] = e.Key
		i++
	}
	return keys
}

func (s *Simple[K, V]) values() []V {
	values := make([]V, s.ll.Len())
	i := 0
	for e := s.ll.Last(); e != nil; e = e.Prev() {
		values[i] = e.Val
		i++
	}
	return values
}
func (s *Simple[K, V]) len() int {
	return s.ll.Len()
}
func (s *Simple[K, V]) cap() int {
	return s.size
}

func (s *Simple[K, V]) purge() {
	for k := range s.items {
		delete(s.items, k)
	}
	s.ll.Init()
}

func (s *Simple[K, V]) resize(size int) int {
	diff := s.Len() - size
	if diff < 0 {
		diff = 0
	}
	for j := 0; j < diff; j++ {
		s.RemoveOldest()
	}
	s.size = size
	return diff
}
