package kit

import (
	"sync"

	"golang.org/x/exp/rand"
)

// CollectionM 是map的集合
type CollectionM[K comparable, V any] struct {
	items map[K]V
}

// MapM 将map中的每个元素通过回调函数进行转换,返回一个V2类型的集合
func MapM[K1 comparable, K2 comparable, V any, V2 any](c *CollectionM[K1, V], fn func(value V, key K1) (K2, V2)) *CollectionM[K2, V2] {
	mapped := make(map[K2]V2, len(c.items))
	for key, value := range c.items {
		mappedKey, mappedValue := fn(value, key)
		mapped[mappedKey] = mappedValue
	}
	return CollectM(mapped)
}

// MapMConc 并发将map中的每个元素通过回调函数进行转换,返回一个V2类型的集合
func MapMConc[K1 comparable, K2 comparable, V any, V2 any](c *CollectionM[K1, V], fn func(value V, key K1) (K2, V2), concurrency ...int) *CollectionM[K2, V2] {
	wg := sync.WaitGroup{}
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	ch := make(chan struct{}, concurrency[0])
	mapped := make(map[K2]V2, len(c.items))
	mu := sync.Mutex{}
	for key, value := range c.items {
		wg.Add(1)
		ch <- struct{}{}
		go func(key K1, value V) {
			defer func() {
				wg.Done()
				<-ch
			}()
			mappedKey, mappedValue := fn(value, key)
			mu.Lock()
			mapped[mappedKey] = mappedValue
			mu.Unlock()
		}(key, value)
	}
	wg.Wait()
	return CollectM(mapped)
}

// CollectM 根据map创建一个集合
func CollectM[K comparable, V any](items map[K]V) *CollectionM[K, V] {
	return &CollectionM[K, V]{items: items}
}

func (c *CollectionM[K, V]) All() map[K]V {
	return c.items
}

// Count 返回集合中元素的数量
func (c *CollectionM[K, V]) Count() int {
	return len(c.items)
}

// IsEmpty 判断集合是否为空
func (c *CollectionM[K, V]) IsEmpty() bool {
	return len(c.items) == 0
}

// IsNotEmpty 判断集合是否不为空
func (c *CollectionM[K, V]) IsNotEmpty() bool {
	return len(c.items) > 0
}

// Keys 返回集合中的所有键
func (c *CollectionM[K, V]) Keys() *CollectionS[K] {
	keys := make([]K, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return CollectS(keys)
}

// Get 获取指定键的值,如果键不存在则返回默认值,如果未提供默认值则返回零值
func (c *CollectionM[K, V]) Get(key K, defaultValue ...V) V {
	if value, ok := c.items[key]; ok {
		return value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return *new(V)
}

// Has 判断集合中是否存在指定键,如果多个键则所有键都存在才返回true
// demo: CollectM(map[string]int{"a": 1, "b": 2}).Has("a", "b") // true
// demo: CollectM(map[string]int{"a": 1, "b": 2}).Has("a", "c") // false
func (c *CollectionM[K, V]) Has(keys ...K) bool {
	for _, key := range keys {
		if _, ok := c.items[key]; !ok {
			return false
		}
	}
	return true
}

// HasAny 判断集合中是否存在指定键中的任意一个,如果多个键则任意一个键存在就返回true
// demo: CollectM(map[string]int{"a": 1, "b": 2}).HasAny("a", "c") // true
// demo: CollectM(map[string]int{"a": 1, "b": 2}).HasAny("c", "d") // false
func (c *CollectionM[K, V]) HasAny(keys ...K) bool {
	for _, key := range keys {
		if _, ok := c.items[key]; ok {
			return true
		}
	}
	return false
}

// Loop 遍历集合中的每个元素并执行给定的回调函数,回调函数无返回值
func (c *CollectionM[K, V]) Loop(fn func(value V, key K)) *CollectionM[K, V] {
	for key, value := range c.items {
		fn(value, key)
	}
	return c
}

// LoopConc 并发遍历集合中的每个元素并执行给定的回调函数,回调函数无返回值
func (c *CollectionM[K, V]) LoopConc(fn func(value V, key K), concurrency ...int) *CollectionM[K, V] {
	wg := sync.WaitGroup{}
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	ch := make(chan struct{}, concurrency[0])
	for key, value := range c.items {
		wg.Add(1)
		ch <- struct{}{}
		go func(key K, value V) {
			defer func() {
				wg.Done()
				<-ch
			}()
			fn(value, key)
		}(key, value)
	}
	wg.Wait()
	return c
}

// Each 遍历集合中的每个元素并执行给定的回调函数
// 根据回调函数的返回值可以中断遍历,不修改集合
// demo: CollectM(map[string]int{"a": 1, "b": 2}).Each(func(value int, key string) bool { fmt.Println(key, value); return true })
// 打印: a 1 b 2
func (c *CollectionM[K, V]) Each(fn func(value V, key K) bool) *CollectionM[K, V] {
	for key, value := range c.items {
		if !fn(value, key) {
			break
		}
	}
	return c
}

// Map 将集合中的每个元素通过回调函数进行转换
// 根据回调函数的返回值生成新的集合,不修改原集合
// demo: CollectM(map[string]int{"a": 1, "b": 2}).Map(func(value int, key string) int { return value * 2 })
// 返回: map[string]int{"a": 2, "b": 4}
func (c *CollectionM[K, V]) Map(fn func(value V, key K) V) *CollectionM[K, V] {
	mapped := make(map[K]V, len(c.items))
	for key, value := range c.items {
		mapped[key] = fn(value, key)
	}
	return CollectM(mapped)
}

// MapConc 并发将集合中的每个元素通过回调函数进行转换
// 根据回调函数的返回值生成新的集合,不修改原集合
// demo: CollectM(map[string]int{"a": 1, "b": 2}).MapConc(func(value int, key string) int { return value * 2 }, 2)
// 返回: map[string]int{"a": 2, "b": 4}
func (c *CollectionM[K, V]) MapConc(fn func(value V, key K) V, concurrency ...int) *CollectionM[K, V] {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	mapped := make(map[K]V, len(c.items))
	ch := make(chan struct{}, concurrency[0])
	for key, value := range c.items {
		wg.Add(1)
		ch <- struct{}{}
		go func(key K, value V) {
			defer func() {
				wg.Done()
				<-ch
			}()
			mu.Lock()
			mapped[key] = fn(value, key)
			mu.Unlock()
		}(key, value)
	}
	wg.Wait()
	return CollectM(mapped)
}

// MapWithKeys 遍历集合并将每个值传递给给定的回调。回调应返回包含单个键/值对的*CollectionM
func (c *CollectionM[K, V]) MapWithKeys(fn func(value V, key K) map[K]V) *CollectionM[K, V] {
	mapped := make(map[K]V, len(c.items))
	for key, value := range c.items {
		for k, v := range fn(value, key) {
			mapped[k] = v
		}
	}
	return CollectM(mapped)
}

// Filter 过滤集合中的元素,满足回调函数条件的元素将被保留
// 根据回调函数的返回值过滤元素,不修改原集合
// demo: CollectM(map[string]int{"a": 1, "b": 2, "c": 3}).Filter(func(value int, key string) bool { return value > 1 })
// 返回: map[string]int{"b": 2, "c": 3}
func (c *CollectionM[K, V]) Filter(fn func(value V, key K) bool) *CollectionM[K, V] {
	filtered := make(map[K]V, len(c.items))
	for key, value := range c.items {
		if fn(value, key) {
			filtered[key] = value
		}
	}
	return CollectM(filtered)
}

// Reject 与Filter相反,过滤掉满足回调函数条件的元素
// demo: CollectM(map[string]int{"a": 1, "b": 2, "c": 3}).Reject(func(value int, key string) bool { return value > 1 })
// 返回: map[string]int{"a": 1}
func (c *CollectionM[K, V]) Reject(fn func(value V, key K) bool) *CollectionM[K, V] {
	rejected := make(map[K]V, len(c.items))
	for key, value := range c.items {
		if !fn(value, key) {
			rejected[key] = value
		}
	}
	return CollectM(rejected)
}

// Values 将集合转换为*CollectionS,index将从0开始
func (c *CollectionM[K, V]) Values() *CollectionS[V] {
	items := make([]V, 0, len(c.items))
	for _, value := range c.items {
		items = append(items, value)
	}
	return CollectS(items)
}

// Chunk 将集合分割成多个指定大小的小集合
// demo: CollectM(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}).Chunk(2)
// 返回: []*CollectionM[string, int]{{"a": 1, "b": 2}, {"c": 3, "d": 4}, {"e": 5}}
func (c *CollectionM[K, V]) Chunk(size int) []*CollectionM[K, V] {
	if size <= 0 {
		return []*CollectionM[K, V]{}
	}
	if len(c.items) == 0 {
		return []*CollectionM[K, V]{}
	}
	chunks := make([]*CollectionM[K, V], 0, (len(c.items)+size-1)/size)
	for k, v := range c.items {
		lastIndex := len(chunks) - 1
		if lastIndex == -1 || len(chunks[lastIndex].items) >= size {
			chunks = append(chunks, CollectM(map[K]V{k: v}))
		} else {
			chunks[lastIndex].items[k] = v
		}
	}
	return chunks
}

// ChunkThenConc 将集合分割成多个指定大小的小集合,并对每个小集合进行并发处理
// 注意: 该函数无返回值
func (c *CollectionM[K, V]) ChunkThenConc(size int, fn func(chunk *CollectionM[K, V]), concurrency ...int) {
	chunks := c.Chunk(size)
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	wg := sync.WaitGroup{}
	ch := make(chan struct{}, concurrency[0])
	for _, chunk := range chunks {
		wg.Add(1)
		ch <- struct{}{}
		go func(chunk *CollectionM[K, V]) {
			defer func() {
				wg.Done()
				<-ch
			}()
			fn(chunk)
		}(chunk)
	}
	wg.Wait()

}

// Merge 合并原始的map
func (c *CollectionM[K, V]) Merge(items map[K]V) *CollectionM[K, V] {
	for k, v := range items {
		c.items[k] = v
	}
	return c
}

// Pluck 集合中的子项需要为结构体或map[K]any, 返回的集合中的元素为子项中指定键的值
// demo: CollectM(map[string]Person{"a":Person{Name: "a", Age: 1}, "b":Person{Name: "b", Age: 2}, "c":Person{Name: "c", Age: 3}}).Pluck("Name")
// 返回: *CollectionS[string]{"a", "b", "c"}
// func (c *CollectionM[K, V]) Pluck(subKey string) *CollectionS[any] {
// 	items := make([]any, 0, len(c.items))
// 	for _, item := range c.items {
// 		itemRef := reflect.ValueOf(item)
// 		if itemRef.Kind() == reflect.Pointer {
// 			if itemRef.IsNil() {
// 				continue
// 			}
// 			itemRef = itemRef.Elem()
// 		}
// 		switch itemRef.Kind() {
// 		case reflect.Struct:
// 			field := itemRef.FieldByName(subKey)
// 			if field.IsValid() && field.CanInterface() {
// 				items = append(items, field.Interface())
// 			}
// 		case reflect.Map:
// 			mapValue := itemRef.MapIndex(reflect.ValueOf(subKey))
// 			if mapValue.IsValid() && mapValue.CanInterface() {
// 				items = append(items, mapValue.Interface())
// 			}
// 		}
// 	}
// 	return CollectS(items)
// }

// Pluck 遍历集合中的每个元素并执行给定的回调函数,返回的集合中的元素为回调函数的返回值
func (c *CollectionM[K, V]) Pluck(fn func(value V, key K) any) *CollectionS[any] {
	items := make([]any, 0, len(c.items))
	for key, value := range c.items {
		items = append(items, fn(value, key))
	}
	return CollectS(items)
}

// Put 将一个键值对添加到集合中
func (c *CollectionM[K, V]) Put(key K, value V) *CollectionM[K, V] {
	c.items[key] = value
	return c
}

// Pull 通过它的键从集合中移除并返回一个项目
func (c *CollectionM[K, V]) Pull(key K) (V, bool) {
	if value, ok := c.items[key]; ok {
		delete(c.items, key)
		return value, true
	}
	var zero V
	return zero, false
}

// Forget 从集合中移除指定键值对
// demo: CollectM(map[string]int{"a": 1, "b": 2}).Forget("a")
// 返回: map[string]int{"b": 2}
func (c *CollectionM[K, V]) Forget(key K) *CollectionM[K, V] {
	delete(c.items, key)
	return c
}

// Random 从集合中返回一个随机键值对
func (c *CollectionM[K, V]) Random() (K, V) {
	randIndex := rand.Intn(len(c.items))
	for key, value := range c.items {
		if randIndex == 0 {
			return key, value
		}
		randIndex--
	}
	var zero K
	var zeroV V
	return zero, zeroV
}

// Randoms 可选参数n,返回集合中的随机键值对,默认返回一个
// demo:
//
//	CollectM(map[string]int{"a": 1, "b": 2, "c": 3}).Randoms(2)
//
// 返回: [{"a": 1}, {"b": 2}]
func (c *CollectionM[K, V]) Randoms(n ...int) *CollectionM[K, V] {
	new := CollectM(map[K]V{})
	if len(n) == 0 {
		k, v := c.Random()
		new.items[k] = v
	} else {
		for i := 0; i < n[0]; i++ {
			k, v := c.Random()
			new.items[k] = v
		}
	}
	return new
}

// When 第一个参数传入为 true 时，将执行给定的成功回调函数。如果为 false 则执行失败回调函数,失败回调函数可选
// 回调函数有两个参数,第一个参数为集合,第二个参数为When传入的第一个参数
// demo:
//
//	CollectM(map[string]int{"a": 1, "b": 2}).When(true, func(c *CollectionM[string, int], cond bool) {
//		fmt.Println(c.All())
//	}).When(false, func(c *CollectionM[string, int], cond bool) {
//		fmt.Println("false")
//	})
func (c *CollectionM[K, V]) When(cond bool, success func(c *CollectionM[K, V], cond bool), fail ...func(c *CollectionM[K, V], cond bool)) *CollectionM[K, V] {
	if cond {
		success(c, cond)
	} else if len(fail) > 0 {
		fail[0](c, cond)
	}
	return c
}

// Reduce 方法将集合减少为单个值，将每次迭代的结果传递给后续迭代
// $carry 在第一次迭代时的值为 null；但是，你可以通过将第二个参数传递给 reduce 来指定其初始值：
// demo:
//
//	CollectM(map[string]int{"a": 1, "b": 2, "c": 3}).Reduce(func(carry int, item int, key string) int {
//			return carry + item
//		}, 0)
//
// 返回: 6
func (c *CollectionM[K, V]) Reduce(fn func(carry V, item V, key K) V, initial ...V) V {
	if len(c.items) == 0 {
		var zero V
		return zero
	}
	var carry V
	if len(initial) > 0 {
		carry = initial[0]
	}
	for key, value := range c.items {
		carry = fn(carry, value, key)
	}
	return carry
}

// Tap 将集合传递给给定的回调函数,回调函数无返回值
// demo:
//
//	CollectM(map[string]int{"a": 1, "b": 2}).Tap(func(c *CollectionM[string, int]) {
//		fmt.Println(c.All())
//	})
func (c *CollectionM[K, V]) Tap(fn func(c *CollectionM[K, V])) *CollectionM[K, V] {
	fn(c)
	return c
}
