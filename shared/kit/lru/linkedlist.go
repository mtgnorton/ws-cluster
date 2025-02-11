package lru

import "fmt"

type Entry[K comparable, V any] struct {
	Key        K
	Val        V
	prev, next *Entry[K, V]
	list       *LinkedList[K, V]
}

func (e *Entry[K, V]) Next() *Entry[K, V] {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

func (e *Entry[K, V]) Prev() *Entry[K, V] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

type LinkedList[K comparable, V any] struct {
	root Entry[K, V]
	len  int
}

func NewLinkedList[K comparable, V any]() *LinkedList[K, V] { return new(LinkedList[K, V]).Init() }

// Init 初始化或清空链表
func (l *LinkedList[K, V]) Init() *LinkedList[K, V] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// Len 返回链表的长度
func (l *LinkedList[K, V]) Len() int { return l.len }

// First 返回链表的第一个元素, 如果链表为空, 返回nil
func (l *LinkedList[K, V]) First() *Entry[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Last 返回链表的最后一个元素, 如果链表为空, 返回nil
func (l *LinkedList[K, V]) Last() *Entry[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// InsertAtHead 将entry插入到链表头部, 并返回entry
func (l *LinkedList[K, V]) InsertAtHead(k K, v V) *Entry[K, V] {
	return l.insertKVAfter(k, v, &l.root)
}

// InsertAtTail 将entry插入到链表尾部, 并返回entry
func (l *LinkedList[K, V]) InsertAtTail(k K, v V) *Entry[K, V] {
	return l.insertKVAfter(k, v, l.root.prev)
}

// RemoveAtHead 删除链表头部的entry, 并返回entry
func (l *LinkedList[K, V]) RemoveAtHead() *Entry[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.RemoveEntry(l.root.next)
}

// RemoveAtTail 删除链表尾部的entry, 并返回entry
func (l *LinkedList[K, V]) RemoveAtTail() *Entry[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.RemoveEntry(l.root.prev)
}

// Find 查找key对应的entry, 如果找到, 返回entry, 否则返回nil
func (l *LinkedList[K, V]) Find(key K) *Entry[K, V] {
	for e := l.First(); e != nil; e = e.Next() {
		if e.Key == key {
			return e
		}
	}
	return nil
}

// Remove 删除key对应的entry, 并返回entry
func (l *LinkedList[K, V]) Remove(key K) *Entry[K, V] {
	if entry := l.Find(key); entry != nil {
		return l.RemoveEntry(entry)
	}
	return nil
}

// RemoveEntry 删除entry, 并返回entry
func (l *LinkedList[K, V]) RemoveEntry(entry *Entry[K, V]) *Entry[K, V] {
	if entry.list != l {
		return nil
	}
	entry.prev.next = entry.next
	entry.next.prev = entry.prev
	entry.next = nil
	entry.prev = nil
	entry.list = nil
	l.len--
	return entry
}

// Move 将entry移动到prevEntry之后
func (l *LinkedList[K, V]) Move(entry, prevEntry *Entry[K, V]) {
	if entry == prevEntry {
		return
	}
	entry.prev.next = entry.next
	entry.next.prev = entry.prev
	entry.prev = prevEntry
	entry.next = prevEntry.next
	entry.prev.next = entry
	entry.next.prev = entry
}

// MoveToFront 将entry移动到链表头部
func (l *LinkedList[K, V]) MoveToFront(entry *Entry[K, V]) {
	if entry.list != l || l.root.next == entry {
		return
	}
	l.Move(entry, &l.root)
}

// IsEmpty 判断链表是否为空
func (l *LinkedList[K, V]) IsEmpty() bool {
	return l.len == 0
}

// String 返回链表的字符串表示
func (l *LinkedList[K, V]) String() string {
	if l.len == 0 {
		return "[]"
	}
	var s string
	for e := l.First(); e != nil; e = e.Next() {
		s += fmt.Sprintf("%v:%v ", e.Key, e.Val)
	}
	return "[" + s[:len(s)-1] + "]"
}

// insertAfter 将entry插入到prevEntry之后, 并返回entry
func (l *LinkedList[K, V]) insertAfter(entry, prevEntry *Entry[K, V]) *Entry[K, V] {
	entry.prev = prevEntry
	entry.next = prevEntry.next
	entry.prev.next = entry
	entry.next.prev = entry
	entry.list = l
	l.len++
	return entry
}

// insertKVAfter key和val生成一个entry, 插入到prevEntry之后, 并返回entry
func (l *LinkedList[K, V]) insertKVAfter(key K, val V, prevEntry *Entry[K, V]) *Entry[K, V] {
	entry := &Entry[K, V]{Key: key, Val: val}
	return l.insertAfter(entry, prevEntry)
}
