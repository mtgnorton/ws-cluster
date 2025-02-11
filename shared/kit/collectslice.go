package kit

import (
	"encoding/json"
	"sort"
	"sync"

	"golang.org/x/exp/rand"
)

// CollectionS 是slice的集合
type CollectionS[T any] struct {
	items []T
}

// MapS 将集合中的每个T类型的元素通过回调函数进行转换,返回一个V类型的集合
// 因为golang方法不支持泛型,所以需要使用函数来实现,但是该函数违反链式调用
// demo:
//
//	MapS(*CollectS([]int{1, 2, 3}), func(item int, index int) string {
//		return fmt.Sprintf("%d", item)
//	})
//
// 返回: *CollectionS[string]{"1", "2", "3"}
func MapS[T, V any](c *CollectionS[T], fn func(item T, index int) V) *CollectionS[V] {
	mapped := make([]V, len(c.items))
	for i, item := range c.items {
		mapped[i] = fn(item, i)
	}
	return CollectS(mapped)
}

// MapSConc 将集合中的每个T类型的元素通过回调函数进行转换,返回一个V类型的集合
// 因为golang方法不支持泛型,所以需要使用函数来实现,但是该函数违反链式调用
func MapSConc[T, V any](c *CollectionS[T], fn func(item T, index int) V, concurrency ...int) *CollectionS[V] {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	mapped := make([]V, len(c.items))
	ch := make(chan struct{}, concurrency[0])
	for i, item := range c.items {
		wg.Add(1)
		ch <- struct{}{}
		go func(i int, item T) {
			defer func() {
				wg.Done()
				<-ch
			}()
			mu.Lock()
			mapped[i] = fn(item, i)
			mu.Unlock()
		}(i, item)
	}
	wg.Wait()
	return CollectS(mapped)
}

// MustAs 将any转换为*CollectionS[V]
// 如果类型断言失败,将抛出panic
// demo:
//
//	MustAs[int](CollectS([]int{1, 2, 3}))
//
// 返回: *CollectionS[int]{1, 2, 3}
func MustAs[V any](c any) *CollectionS[V] {
	newCollection := &CollectionS[V]{}
	if collection, ok := c.(*CollectionS[any]); ok {
		for _, item := range collection.items {
			if _, ok := item.(V); !ok {
				panic("As: 类型断言失败")
			}
			newCollection.items = append(newCollection.items, item.(V))
		}
		return newCollection
	}
	panic("As: 类型断言失败")
}

// Flatten 将多维Slice转为一维Slice
// demo:
//
//	CollectS([][]int{{1, 2}, {3, 4}}).Flatten(func(item []int) []int {
//		return item
//	})
//
// 返回: *CollectionS[int]{1, 2, 3, 4}
func Flatten[T, V any](s *CollectionS[T], fn func(item T) []V) *CollectionS[V] {
	items := make([]V, 0)
	for _, item := range s.items {
		items = append(items, fn(item)...)
	}
	return CollectS(items)
}

// CollectS 根据slice创建一个集合
// demo:
//
//	CollectS([]int{1, 2, 3})
//
// 返回: *CollectionS[int]{1, 2, 3}
func CollectS[T any](items []T) *CollectionS[T] {
	return &CollectionS[T]{items: items}
}

// All 返回集合中的所有元素
// demo:
//
//	CollectS([]int{1, 2, 3}).All()
//
// 返回: []int{1, 2, 3}
func (c *CollectionS[T]) All() []T {
	return c.items
}

// Count 返回集合中元素的数量
// demo:
//
//	CollectS([]int{1, 2, 3}).Count()
//
// 返回: 3
func (c *CollectionS[T]) Count() int {
	return len(c.items)
}

// IsEmpty 判断集合是否为空
// demo:
//
//	CollectS([]int{}).IsEmpty()
//
// 返回: true
func (c *CollectionS[T]) IsEmpty() bool {
	return len(c.items) == 0
}

// IsNotEmpty 判断集合是否不为空
// demo:
//
//	CollectS([]int{1, 2, 3}).IsNotEmpty()
//
// 返回: true
func (c *CollectionS[T]) IsNotEmpty() bool {
	return len(c.items) > 0
}

// Loop 遍历集合中的每个元素并执行给定的回调函数,回调函数无返回值,返回原集合
func (c *CollectionS[T]) Loop(fn func(item T, index int)) *CollectionS[T] {
	for i, item := range c.items {
		fn(item, i)
	}
	return c
}

// LoopConc 并发遍历集合中的每个元素并执行给定的回调函数,回调函数无返回值,返回原集合
func (c *CollectionS[T]) LoopConc(fn func(item T, index int), concurrency ...int) *CollectionS[T] {
	wg := sync.WaitGroup{}
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	ch := make(chan struct{}, concurrency[0])
	for i, item := range c.items {
		wg.Add(1)
		ch <- struct{}{}
		go func(i int, item T) {
			defer func() {
				wg.Done()
				<-ch
			}()
			fn(item, i)
		}(i, item)
	}
	wg.Wait()
	return c
}

// Each 遍历集合中的每个元素并执行给定的回调函数,根据
// 根据回调函数的返回值可以中断遍历,不修改集合
// demo:
//
//	CollectS([]int{1, 2, 3}).Each(func(item int, index int) bool {
//		fmt.Println(item)
//		return true
//	})
//
// 打印: 1 2 3
func (c *CollectionS[T]) Each(fn func(item T, index int) bool) *CollectionS[T] {
	for i, item := range c.items {
		if !fn(item, i) {
			break
		}
	}
	return c
}

// Map 将集合中的每个元素通过回调函数进行转换
// 根据回调函数的返回值生成新的集合,不修改原集合
// demo:
//
//	CollectS([]int{1, 2, 3}).Map(func(item int, index int) int {
//		return item * 2
//	})
//
// 返回: []int{2, 4, 6}
func (c *CollectionS[T]) Map(fn func(item T, index int) T) *CollectionS[T] {
	mapped := make([]T, len(c.items))
	for i, item := range c.items {
		mapped[i] = fn(item, i)
	}
	return CollectS(mapped)
}

// MapAny 将集合中的每个元素通过回调函数进行转换,回调函数返回值为any
// demo:
//
//	CollectS([]int{1, 2, 3}).MapAny(func(item int, index int) any {
//		return fmt.Sprintf("%d", item)
//	})
//
// 返回: *CollectionS[any]{"1", "2", "3"}
func (c *CollectionS[T]) MapAny(fn func(item T, index int) any) *CollectionS[any] {
	mapped := make([]any, len(c.items))
	for i, item := range c.items {
		mapped[i] = fn(item, i)
	}
	return CollectS(mapped)
}

// MapConc 并发将集合中的每个元素通过回调函数进行转换
// demo:
//
//	CollectS([]int{1, 2, 3}).MapConc(func(item int, index int) int {
//		return item * 2
//	}, 2)
//
// 返回: []int{2, 4, 6}
func (c *CollectionS[T]) MapConc(fn func(item T, index int) T, concurrency ...int) *CollectionS[T] {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	mapped := make([]T, len(c.items))
	ch := make(chan struct{}, concurrency[0])
	for i, item := range c.items {
		wg.Add(1)
		ch <- struct{}{}
		go func(i int, item T) {
			defer func() {
				wg.Done()
				<-ch
			}()
			mu.Lock()
			mapped[i] = fn(item, i)
			mu.Unlock()
		}(i, item)
	}
	wg.Wait()
	return CollectS(mapped)
}

// MapConcAny 并发将集合中的每个元素通过回调函数进行转换,回调函数返回值为any
func (c *CollectionS[T]) MapConcAny(fn func(item T, index int) any, concurrency ...int) *CollectionS[any] {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	mapped := make([]any, len(c.items))
	ch := make(chan struct{}, concurrency[0])
	for i, item := range c.items {
		wg.Add(1)
		ch <- struct{}{}
		go func(i int, item T) {
			defer func() {
				wg.Done()
				<-ch
			}()
			mu.Lock()
			mapped[i] = fn(item, i)
			mu.Unlock()
		}(i, item)
	}
	wg.Wait()
	return CollectS(mapped)
}

// MapWithKeys 遍历集合并将每个值传递给给定的回调。回调应返回包含单个键/值对的*CollectionM
// 常用来将slice转换为map
// demo:
//
//	CollectS([]int{1, 2, 3}).MapWithKeys(func(item int, index int) map[string]int {
//		return map[string]int{"a": item}
//	})
//
// 返回: *CollectionM[string, int]{{"a": 1, "b": 2, "c": 3}}
func (c *CollectionS[T]) MapWithKeys(fn func(item T, index int) map[string]T) *CollectionM[string, T] {
	mapped := make(map[string]T)
	for i, item := range c.items {
		for k, v := range fn(item, i) {
			mapped[k] = v
		}
	}
	return CollectM(mapped)
}

// Sum 根据回调函数返回值计算集合中元素的和
// demo:
//
//	CollectS([]int{1, 2, 3}).Sum(func(item int, index int) int {
//		return item * 2
//	})
//
// 返回: 12
func (c *CollectionS[T]) Sum(fn func(item T, index int) float64) float64 {
	var sum float64
	for i, item := range c.items {
		sum += fn(item, i)
	}
	return sum
}

// Chunk 将集合分割成多个指定大小的小集合
// demo:
//
//	CollectS([]int{1, 2, 3, 4, 5}).Chunk(2)
//
// 返回: []*CollectionS[int]{{1, 2}, {3, 4}, {5}}
func (c *CollectionS[T]) Chunk(size int) []*CollectionS[T] {
	if size <= 0 {
		return []*CollectionS[T]{}
	}
	if len(c.items) == 0 {
		return []*CollectionS[T]{}
	}
	var chunks []*CollectionS[T]
	for i := 0; i < len(c.items); i += size {
		end := i + size
		if end > len(c.items) {
			end = len(c.items)
		}
		chunks = append(chunks, CollectS(c.items[i:end]))
	}
	return chunks
}

// ChunkThenConc 先将集合分割成多个小集合,然后并发处理小集合
// 注意: 该函数无返回值
// demo:
//
//	CollectS([]int{1, 2, 3, 4, 5}).ChunkThenConc(2, func(chunk *CollectionS[int]) {
//		fmt.Println(chunk.All())
//	}, 2)
//
// 打印:
// [1 2]
// [3 4]
// [5]
func (c *CollectionS[T]) ChunkThenConc(size int, fn func(chunk *CollectionS[T]), concurrent ...int) {
	chunks := c.Chunk(size)
	if len(concurrent) == 0 {
		concurrent = []int{1}
	}
	wg := sync.WaitGroup{}
	ch := make(chan struct{}, concurrent[0])
	for _, chunk := range chunks {
		wg.Add(1)
		ch <- struct{}{}
		go func(chunk *CollectionS[T]) {
			defer func() {
				wg.Done()
				<-ch
			}()
			fn(chunk)
		}(chunk)
	}
	wg.Wait()
}

// Every 检查集合中的每个元素是否都通过给定的真值测试
// demo:
//
//	CollectS([]int{2, 4, 6}).Every(func(item int, index int) bool {
//		return item%2 == 0
//	})
//
// 返回: true
func (c *CollectionS[T]) Every(fn func(item T, index int) bool) bool {
	for i, item := range c.items {
		if !fn(item, i) {
			return false
		}
	}
	return true
}

// Filter 根据给定的回调函数过滤集合中的元素,满足回调函数条件的元素将被保留
// 示例:
//
//	CollectS([]int{1, 2, 3, 4, 5}).Filter(func(item int, index int) bool {
//		return item > 2
//	})
//
// 返回: []int{3, 4, 5}
func (c *CollectionS[T]) Filter(fn func(item T, index int) bool) *CollectionS[T] {
	var filtered []T
	for i, item := range c.items {
		if fn(item, i) {
			filtered = append(filtered, item)
		}
	}
	return CollectS(filtered)
}

// Reject 与Filter相反,过滤掉满足回调函数条件的元素
// demo:
//
//	CollectS([]int{1, 2, 3, 4, 5}).Reject(func(item int, index int) bool {
//		return item > 2
//	})
//
// 返回: []int{1, 2}
func (c *CollectionS[T]) Reject(fn func(item T, index int) bool) *CollectionS[T] {
	var rejected []T
	for i, item := range c.items {
		if !fn(item, i) {
			rejected = append(rejected, item)
		}
	}
	return CollectS(rejected)
}

// First 返回集合中通过给定真值测试的第一个元素
// demo:
//
//	CollectS([]int{1, 2, 3, 4, 5}).First(func(item int, index int) bool {
//		return item > 3
//	})
//
// 返回: 4, true
func (c *CollectionS[T]) First(fn ...func(item T, index int) bool) (T, bool) {
	if len(c.items) == 0 {
		var zero T
		return zero, false
	}

	if len(fn) == 0 {
		return c.items[0], true
	}

	for i, item := range c.items {
		if fn[0](item, i) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// Merge 合并原始slice
// demo:
//
//	CollectS([]int{1, 2}).Merge([]int{3, 4})
//
// 返回: *CollectionS[int]{1, 2, 3, 4}
func (c *CollectionS[T]) Merge(items []T) *CollectionS[T] {
	return CollectS(append(c.items, items...))
}

// Pluck 遍历集合中的每个元素并执行给定的回调函数,返回的集合中的元素为回调函数的返回值
// demo:
//
//	CollectS([]int{1, 2, 3}).Pluck(func(item int, index int) any {
//		return item * 2
//	})
//
// 返回: *CollectionS[any]{2, 4, 6}
func (c *CollectionS[T]) Pluck(fn func(item T, index int) any) *CollectionS[any] {
	items := make([]any, len(c.items))
	for i, item := range c.items {
		items[i] = fn(item, i)
	}
	return CollectS(items)
}

// Shift 移除并返回集合中的第一个元素
// demo:
//
//	CollectS([]int{1, 2, 3}).Shift()
//
// 返回: 1, true
func (c *CollectionS[T]) Shift() (T, bool) {
	if len(c.items) == 0 {
		var zero T
		return zero, false
	}
	first := c.items[0]
	c.items = c.items[1:]
	return first, true
}

// Pop 移除并返回集合中的最后一个元素
// demo:
//
//	CollectS([]int{1, 2, 3}).Pop()
//
// 返回: 3, true
func (c *CollectionS[T]) Pop() (T, bool) {
	if len(c.items) == 0 {
		var zero T
		return zero, false
	}
	last := c.items[len(c.items)-1]
	c.items = c.items[:len(c.items)-1]
	return last, true
}

// Prepend 在集合的开头添加一个元素
// demo:
//
//	CollectS([]int{2, 3}).Prepend(1)
//
// 返回: *CollectionS[int]{1, 2, 3}
func (c *CollectionS[T]) Prepend(item T) *CollectionS[T] {
	return CollectS(append([]T{item}, c.items...))
}

// Push 在集合的末尾添加一个元素
// demo:
//
//	CollectS([]int{1, 2}).Push(3)
//
// 返回: *CollectionS[int]{1, 2, 3}
func (c *CollectionS[T]) Push(item T) *CollectionS[T] {
	return CollectS(append(c.items, item))
}

// Random 从集合中返回一个随机元素
// demo:
//
//	CollectS([]int{1, 2, 3}).Random()
//
// 返回: 随机返回1、2或3中的一个
func (c *CollectionS[T]) Random() T {
	return c.items[rand.Intn(len(c.items))]
}

// Range 返回一个包含指定范围之间整数的集合
// demo:
//
//	CollectS([]int{}).Range(1, 5)
//
// 返回: *CollectionS[int]{1, 2, 3, 4, 5}
func (c *CollectionS[T]) Range(start, end int) *CollectionS[int] {
	items := make([]int, end-start+1)
	for i := range items {
		items[i] = start + i
	}
	return CollectS(items)
}

// Reverse 反转集合中的元素
// demo:
//
//	CollectS([]int{1, 2, 3}).Reverse()
//
// 返回: *CollectionS[int]{3, 2, 1}
func (c *CollectionS[T]) Reverse() *CollectionS[T] {
	for i, j := 0, len(c.items)-1; i < j; i, j = i+1, j-1 {
		c.items[i], c.items[j] = c.items[j], c.items[i]
	}
	return c
}

// Reduce 方法将集合减少为单个值，将每次迭代的结果传递给后续迭代
// $carry 在第一次迭代时的值为 null；但是，你可以通过将第二个参数传递给 reduce 来指定其初始值：
// demo:
//
//	CollectS([]int{1, 2, 3}).Reduce(func(carry int, item int, index int) int {
//		return carry + item
//	}, 0)
//
// 返回: 6
func (c *CollectionS[T]) Reduce(fn func(carry T, item T, index int) T, initial ...T) T {
	if len(c.items) == 0 {
		var zero T
		return zero
	}
	var carry T
	if len(initial) > 0 {
		carry = initial[0]
	}
	for i := 0; i < len(c.items); i++ {
		carry = fn(carry, c.items[i], i)
	}
	return carry
}

// Skip 跳过集合中的前n个元素
// demo:
//
//	CollectS([]int{1, 2, 3}).Skip(2)
//
// 返回: *CollectionS[int]{3}
func (c *CollectionS[T]) Skip(n int) *CollectionS[T] {
	return CollectS(c.items[n:])
}

// Take 方法返回一个具有指定数量项目的新集合
// 也可以传递一个负整数以从集合末尾获取指定数量的项目
// demo:
//
//	CollectS([]int{1, 2, 3}).Take(2)
//
// 返回: *CollectionS[int]{1, 2}
// CollectS([]int{1, 2, 3}).Take(-1)
//
// 返回: *CollectionS[int]{3}

func (c *CollectionS[T]) Take(n int) *CollectionS[T] {
	if n == 0 {
		return CollectS([]T{})
	}
	if n < 0 {
		return CollectS(c.items[len(c.items)+n:])
	}
	return CollectS(c.items[:n])
}

// When 第一个参数传入为 true 时，将执行给定的成功回调函数。如果为 false 则执行失败回调函数,失败回调函数可选
// 回调函数有两个参数,第一个参数为集合,第二个参数为When传入的第一个参数
// demo:
//
//	CollectS([]int{1, 2, 3}).When(true, func(c *CollectionS[int], cond bool) {
//		fmt.Println(c.All())
//	}).When(false, func(c *CollectionS[int], cond bool) {
//		fmt.Println("false")
//	})
func (c *CollectionS[T]) When(cond bool, success func(c *CollectionS[T], cond bool), fail ...func(c *CollectionS[T], cond bool)) *CollectionS[T] {
	if cond {
		success(c, cond)
	} else if len(fail) > 0 {
		fail[0](c, cond)
	}
	return c
}

// Sort 根据回调函数对集合中的元素进行排序
// demo:
//
//	CollectS([]int{3, 1, 2}).Sort(func(a, b int) bool {
//		return a < b // 升序
//	})
//
// 返回: *CollectionS[int]{1, 2, 3}
func (c *CollectionS[T]) Sort(fn func(a, b T) bool) *CollectionS[T] {
	sort.Slice(c.items, func(i, j int) bool {
		return fn(c.items[i], c.items[j])
	})
	return c
}

// Tap 将集合传递给给定的回调函数,并返回集合本身
// demo:
//
//	CollectS([]int{1, 2, 3}).Tap(func(c *CollectionS[int]) {
//		fmt.Println(c.All())
//	})
func (c *CollectionS[T]) Tap(fn func(c *CollectionS[T])) *CollectionS[T] {
	fn(c)
	return c
}

// Zip 将集合中的元素与另一个slice中的元素一一对应
// demo:
//
//	CollectS([]int{1, 2, 3}).Zip([]string{"a", "b", "c"})
//
// 返回: *CollectionS[[]any]{{1, "a"}, {2, "b"}, {3, "c"}}
func (c *CollectionS[T]) Zip(items []any) *CollectionS[[]any] {
	zipped := make([][]any, len(c.items))
	for i := range c.items {
		zipped[i] = []any{c.items[i], items[i]}
	}
	return CollectS(zipped)
}

// Unique 根据回调函数移除集合中的重复元素
func (c *CollectionS[T]) Unique(fn func(item T, index int) any) *CollectionS[T] {
	mapping := make(map[any]struct{})
	newCollection := make([]T, 0)
	for i, item := range c.items {
		v := fn(item, i)
		if _, ok := mapping[v]; !ok {
			mapping[v] = struct{}{}
			newCollection = append(newCollection, item)
		}
	}
	return CollectS(newCollection)
}

// Head 可选参数n,返回集合中的前n个元素,默认返回第一个元素
// demo:
//
//	CollectS([]int{1, 2, 3}).Head()
//
// 返回: 1
func (c *CollectionS[T]) Head(n ...int) *CollectionS[T] {
	if len(n) == 0 {
		return CollectS([]T{c.items[0]})
	}
	return CollectS(c.items[:n[0]])
}

// Tail 可选参数n,返回集合中的后n个元素,默认返回最后一个元素
// demo:
//
//	CollectS([]int{1, 2, 3}).Tail()
//
// 返回: 3
func (c *CollectionS[T]) Tail(n ...int) *CollectionS[T] {
	if len(n) == 0 {
		return CollectS([]T{c.items[len(c.items)-1]})
	}
	return CollectS(c.items[len(c.items)-n[0]:])
}

// Json 将集合转换为json字符串
// Demo:
//
//	CollectS([]int{1, 2, 3}).Json()
//
// 返回: [1,2,3]
func (c *CollectionS[T]) Json() ([]byte, error) {
	return json.Marshal(c.items)
}
