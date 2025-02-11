package lru

import (
	"fmt"
	"sync"
)

// ARC （Adaptive Replacement Cache）：自适应缓存替换算法
// . . . [   B1  <-[     T1    <-!->      T2   ]->  B2   ] . .
//
//		 [ . . . . [ . . . . . . ! . .^. . . . ] . . . . ]
//	          	   [   fixed cache size (c)    ]
//
// ^表示目标大小，!表示实际大小
// 最内部的[],代表实际的缓存,大小固定,但是记录可以在b1,b2 中移动
// L1 从右到左表示从新到旧,^可能等于,大于，小于!
// 新记录首先进入t1,并逐渐被推向到左侧,然后会从t1->b1,最后完全被淘汰
// L1 中的任何记录再次被访问,会进入到L2,在这里，会逐渐被推向右侧,然后会从t2->b2,l2中再次命中的记录可以无限重复此操作,知道在b2的最右侧消失
//
// b1和b2 的作用
// 记录首次或者再次进入t1或t2,将导致!向^移动，如果缓存中不存在可用空间,则该标记还确定 t1 或 t2 是否将逐出条目。
// 命中b1将增加t1的大小,将^向右移动,t2中的最后一个条目将被淘汰到b2
// 命中b2将缩小t1的大小,将^向左移动,t1中的最后一个条目将被淘汰到b1
// 缓存未命中不会影响^,但会影响!边界靠近^
//
// size 代表缓存大小
// t1 是最近访问的缓存
// t2 是最频繁访问的缓存,至少引用过两次
// b1 是最近从 t1 中淘汰的缓存
// b2 是最近从 t2 中淘汰的缓存
// t1+b1 合称L1，t2+b2 合称L2
// p 如果p值增大,t1 的大小会增大,t2 的大小会减小,反之亦然,b1 的大小会增大,b2 的大小会增大,b1的大小会减小
// 添加和查找只针对t1和t2,当新元素较多的时候,t1长度会增长,t2长度会缩小,而当旧元素命中较多时候 t2长度会增长t1长度会缩小。
//
// 替换过程
// 新元素: 添加新元素到T1,若空间不足,淘汰T2
// 已存在元素: 若在B1或B2存在,移动到T2,若空间不足,淘汰T1
// 查询命中: 若T1查询命中,移动到T2
type ARC[K comparable, V any] struct {
	size    int
	p       int
	t1, t2  *Simple[K, V]
	b1, b2  *Simple[K, struct{}]
	mu      sync.Mutex
	evictCB func(key K, value V)
}

type HitType int

const (
	HitB1 HitType = iota + 1
	HitB2

	HitUnknown
)

func NewMustARC[K comparable, V any](size int, cbs ...func(key K, value V)) *ARC[K, V] {
	arc, err := NewARC[K, V](size, cbs...)
	if err != nil {
		panic(err)
	}
	return arc
}

func NewARC[K comparable, V any](size int, cbs ...func(key K, value V)) (*ARC[K, V], error) {
	t1, err := NewSimple[K, V](size)
	if err != nil {
		return nil, err
	}
	t2, err := NewSimple[K, V](size)
	if err != nil {
		return nil, err
	}
	b1, err := NewSimple[K, struct{}](size)
	if err != nil {
		return nil, err
	}
	b2, err := NewSimple[K, struct{}](size)
	if err != nil {
		return nil, err
	}
	var cb func(key K, value V)
	if len(cbs) > 0 {
		cb = cbs[0]
	}
	return &ARC[K, V]{
		size:    size,
		t1:      t1,
		t2:      t2,
		b1:      b1,
		b2:      b2,
		evictCB: cb,
	}, nil
}

// Add
// 当记录移动到影子缓存时evict返回true
// 为什么在命中 b1 和 b2 时调整 p：
// 命中 b1 表示 T1 不足够大:
// 当一个请求的数据在 b1 中命中时，表明该数据最近曾被替换出 T1。这个命中提示我们，之前这部分数据被认为是“不够频繁使用”的，但现在它又被访问了。这样的命中表明 T1 可能设置得过小，没有足够的空间来容纳这些最近使用的数据。因此，我们需要增加 T1 的大小，这通过增加 p 的值来实现。
//
// 命中 b2 表示 T2 不足够大:
// 同样地，当一个请求的数据在 b2 中命中时，这表明该数据最近曾被替换出 T2。这个命中提示我们，之前这部分数据被认为是“频繁使用”的，但是因为 T2 空间不足而被替换。这样的命中表明 T2 可能设置得过小，没有足够的空间来容纳这些频繁使用的数据。因此，我们需要增加 T2 的大小，这可以通过减少 p 的值来实现。
func (a *ARC[K, V]) Add(key K, value V) (evict bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.t1.contains(key) { // 如果在t1存在，移动到t2
		a.t1.remove(key)
		a.t2.add(key, value)
		return
	} else if a.t2.contains(key) { // 如果在t2存在，移动到首部
		a.t2.add(key, value)
		return
	} else if a.b1.contains(key) { // 如果在b1存在，移动到t2
		a.adjustP(HitB1)
		evict = a.moveToGhost(HitB1)
		a.b1.remove(key)
		a.t2.add(key, value)
		return
	} else if a.b2.contains(key) { // 如果在b2存在，移动到t2,增大p值，
		a.adjustP(HitB2)
		evict = a.moveToGhost(HitB2)
		a.b2.remove(key)
		a.t2.add(key, value)
		return
	}
	evict = a.moveToGhost(HitUnknown)
	// a.size-a.p 代表b1的最大大小,假设size为100,p为20,则b1的最大大小为80
	if a.b1.Len() > a.size-a.p {
		a.b1.removeOldest()
	}
	// a.p 代表b2的最大大小,假设size为100,p为20,则b2的最大大小为20
	if a.b2.Len() > a.p {
		a.b2.removeOldest()
	}
	a.t1.add(key, value)
	return
}

// adjustP 动态调整p,当b1命中记录时，说明t1可能太小，适当增大p
// 当b2命中记录时，说明t2有可能太小，适当减小p
// 如果p值增大,t1的大小会增大,b1的大小会减小,t2的大小会减小,b2的大小会增大
func (a *ARC[K, V]) adjustP(hitType HitType) int {
	var (
		delta = 1
		b1Len = a.b1.len()
		b2Len = a.b2.len()
	)
	if hitType == HitB1 { // t1 太小,适当的增大p
		if b2Len > b1Len {
			delta = b2Len / b1Len
		}
		if a.p+delta >= a.size {
			a.p = a.size
		} else {
			a.p += delta
		}
	} else { // t2 太小,适当的减小p
		if b1Len > b2Len {
			delta = b1Len / b2Len
		}
		if a.p-delta < 0 {
			a.p = 0
		} else {
			a.p -= delta
		}
	}
	return a.p
}

// moveToGhost 移动到影子缓存 b1 或 b2
// 缓存已满时，决定是从 T1 还是 T2 中逐出条目。选择哪个列表逐出条目取决于 p 的当前值，p 值表示 T1 和 T2 的目标大小。
func (a *ARC[K, V]) moveToGhost(hitType HitType) (evict bool) {
	var (
		t1Len = a.t1.len()
		t2Len = a.t2.len()
	)
	if t1Len+t2Len < a.size {
		return
	}
	if t1Len > 0 && (t1Len > a.p || (t1Len == a.p && hitType == HitB2)) {
		if k, v, ok := a.t1.removeOldest(); ok {
			a.evictCB(k, v)
			a.b1.add(k, struct{}{})
			return true
		}
	} else {
		if k, v, ok := a.t2.removeOldest(); ok {
			a.evictCB(k, v)
			a.b2.add(k, struct{}{})
			return true
		}
	}
	return false
}

func (a *ARC[K, V]) Get(key K) (value V, ok bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if value, ok = a.t1.peek(key); ok {
		a.t1.remove(key)
		a.t2.add(key, value)
		return
	} else if value, ok = a.t2.get(key); ok {
		return
	}
	return
}

func (a *ARC[K, V]) Contains(key K) (ok bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.t1.contains(key) || a.t2.contains(key)
}

func (a *ARC[K, V]) Peek(key K) (value V, ok bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if value, ok = a.t1.peek(key); ok {
		return
	}
	return a.t2.peek(key)
}

func (a *ARC[K, V]) Remove(key K) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.t1.remove(key) {
		return true
	}
	if a.t2.remove(key) {
		return true
	}
	if a.b1.remove(key) {
		return true
	}
	if a.b2.remove(key) {
		return true
	}
	return false

}

func (a *ARC[K, V]) Keys() []K {
	a.mu.Lock()
	defer a.mu.Unlock()
	k1 := a.t1.keys()
	k2 := a.t2.keys()
	return append(k1, k2...)
}

func (a *ARC[K, V]) Values() []V {
	a.mu.Lock()
	defer a.mu.Unlock()
	v1 := a.t1.values()
	v2 := a.t2.values()
	return append(v1, v2...)
}

func (a *ARC[K, V]) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.t1.len() + a.t2.len()
}

func (a *ARC[K, V]) Cap() int {
	return a.size
}

func (a *ARC[K, V]) Purge() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.t1.purge()
	a.t2.purge()
	a.b1.purge()
	a.b2.purge()
}

func (a *ARC[K, V]) String() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return fmt.Sprintf("ARC{size=%d,len:%d, p=%d,  \n t1=%s, \n t2=%s, \n b1=%s, \n b2=%s}", a.size, a.t1.len()+a.t2.len(), a.p, a.t1, a.t2, a.b1, a.b2)
}
