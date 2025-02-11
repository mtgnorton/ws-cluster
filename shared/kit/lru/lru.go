package lru

// LRU
// 从时间角度对缓存条目进行淘汰，即最长时间没有被使用的缓存条目会被淘汰。该算法有一个问题：如果某些历史数据突然被大量访问，但仅仅访问一次，就可能会把那些需要频繁访问的缓存条目给淘汰掉，造成之后大量频繁访问的缓存条目出现 cache-miss。

// LFU
// 最不常用算法。从访问频率角度对缓存条目进行淘汰，即访问频率最少的缓存条目会被淘汰。该算法也存在问题：如果之前频繁访问过一些缓存条目，但是现在并不会访问这些条目，这些条目也会一直占据缓冲区，很难被淘汰。

// LRU-K
// 相比于 LRU， LRU-K 算法多维护一个队列，用来记录所有缓存数据被访问的历史，只有当数据访问的次数达到 K 时，才将数据放入真正的缓存队列。整个缓存运作过程如下：
//	数据第一次被访问，直接进入访问历史队列
//	访问历史队列也是按照 LRU 的规则进行淘汰
//	如果历史队列中的数据访问达到 K 次后，将数据从历史队列中删除，移入到缓存队列中
//	缓存数据队列仍然按照 LRU 的规则进行淘汰
//	LRU-K 一定程度上解决了 LRU 的缺点。实际应用中，通常采用 LRU-2。

// 2Q
// 2Q 即 two-queues 算法，类似于 LRU-2，也是使用两个缓存队列，只不过一个是 FIFO 队列，一个是 LRU 队列。缓存的运作过程如下：
//	数据第一次访问，插入到 FIFO 队列
//	如果 FIFO 队列中的数据再次被访问，将移入到 LRU 队列
//	FIFO 按照先进先出的方式进行数据淘汰
//	LRU 队列按照 LRU 规则进行数据淘汰

// ARC
// ARC（Adaptive Replacement Cache）：自适应缓存替换算法。它同时结合了 LRU 和 LFU，当访问的数据趋向于最近访问的条目时，会更多地命中 LRU cache；当访问的数据趋向于最频繁的条目时，会更多地命中 LFU cache。ARC 会动态调整 LRU 和 LFU 的比例，从而提高缓存命中率。

type LRU[K comparable, V any] interface {
	// 添加一个值到缓存，如果发生了淘汰则返回 true，更新最近使用。
	Add(key K, value V) bool

	// 返回 key 的值，更新最近使用。
	Get(key K) (value V, ok bool)

	// 检查 key 是否存在于缓存中，不更新最近使用。
	Contains(key K) (ok bool)

	// 返回 key 的值，不更新最近使用。
	Peek(key K) (value V, ok bool)

	// 从缓存中移除 key。
	Remove(key K) (isFound bool)

	//// 移除缓存中最老的条目。
	//RemoveOldest() (key K, value V, isFound bool)
	//
	//// 返回缓存中最老的条目。#key, value, isFound
	//GetOldest() (key K, value V, isFound bool)
	//
	//GetNewest() (key K, value V, isFound bool)

	// 返回缓存中的所有 key 的切片，从最老到最新。
	Keys() []K

	// 返回缓存中的所有 value 的切片，从最老到最新。
	Values() []V

	// 返回缓存中的条目数量。
	Len() int

	// 返回缓存的容量。
	Cap() int

	//// 清空所有缓存条目。
	//Purge()
	//
	//// 调整缓存大小，返回被淘汰的数量。
	//Resize(int) int
}
