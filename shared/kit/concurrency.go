package kit

import (
	"context"
	"sync"
)

type Result[T any, V any] struct {
	Key    int
	Item   T
	Result V
	Error  error
}

// AsyncOne 并发执行fn,拿到第一个结果就返回
func AsyncOne[T any, V any](s []T, fn func(T) (V, error), concurrency ...int) (result Result[T, V]) {
	rsCh, cancel := Async(s, fn, concurrency...)
	defer cancel()
	for i := 0; i < len(s); i++ {
		rs := <-rsCh
		if rs.Error == nil {
			return rs
		}
		result = rs
	}
	return
}

// Async 并发执行fn，通过rsCh返回结果
func Async[T any, V any](s []T, fn func(T) (V, error), concurrency ...int) (<-chan Result[T, V], func()) {
	conc := 1
	if len(concurrency) > 0 && concurrency[0] > 0 {
		conc = concurrency[0]
	}
	ch := make(chan struct{}, conc)
	resultCh := make(chan Result[T, V], len(s))
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	go func(s []T) {
		for i, item := range s {
			wg.Add(1)
			select {
			case <-ctx.Done():
				return
			default:
				select {
				case <-ctx.Done():
					return
				case ch <- struct{}{}:
				}
			}
			// select {
			// case <-ctx.Done():
			// 	return
			// case ch <- struct{}{}:
			// }
			go func(item T, index int) {
				defer func() {
					wg.Done()
					<-ch
				}()
				v, err := fn(item)
				select {
				case <-ctx.Done():
					return
				default:
					select {
					case <-ctx.Done():
						return
					case resultCh <- Result[T, V]{
						Key:    index,
						Item:   item,
						Result: v,
						Error:  err,
					}:
					}
				}

			}(item, i)
		}
		wg.Wait()
	}(s)

	return resultCh, func() {
		cancel()
		close(ch)
		close(resultCh)
	}
}

// LoopSliceConcurrent 并发遍历slice
func LoopSliceConcurrent[T any](s []T, fn func(index int, item T), concurrency ...int) {
	wg := sync.WaitGroup{}
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	ch := make(chan struct{}, concurrency[0])
	for i, item := range s {
		wg.Add(1)
		ch <- struct{}{}
		go func(i int, item T) {
			defer func() {
				wg.Done()
				<-ch
			}()
			fn(i, item)
		}(i, item)
	}
	wg.Wait()
}

// LoopMapConcurrent 并发遍历map
func LoopMapConcurrent[K comparable, V any](m map[K]V, fn func(key K, value V), concurrency ...int) {
	wg := sync.WaitGroup{}
	if len(concurrency) == 0 {
		concurrency = []int{1}
	}
	ch := make(chan struct{}, concurrency[0])
	for key, value := range m {
		wg.Add(1)
		ch <- struct{}{}
		go func(key K, value V) {
			defer func() {
				wg.Done()
				<-ch
			}()
			fn(key, value)
		}(key, value)
	}
	wg.Wait()
}

// ChunkSlice 将slice分割为多个size大小的slice,然后并发concurrent处理
// 如果size<=0,则返回空slice
// 如果concurrent<=0,则并发数为1
func ChunkSlice[T any](s []T, size int, fn func(chunk []T), concurrent ...int) {
	if len(concurrent) == 0 {
		concurrent = []int{1}
	}
	if size <= 0 {
		return
	}
	if len(s) == 0 {
		return
	}
	length := len(s)

	wg := sync.WaitGroup{}
	ch := make(chan struct{}, concurrent[0])

	for i := 0; i < length; i += size {
		end := Min(i+size, length)
		chunk := s[i:end]

		wg.Add(1)
		ch <- struct{}{}
		go func(chunk []T) {
			defer func() {
				wg.Done()
				<-ch
			}()
			fn(chunk)
		}(chunk)
	}
	wg.Wait()
}

// ChunkMap 将map分割为多个size大小的map,然后并发concurrent处理
// 如果size<=0,则返回空map
// 如果concurrent<=0,则并发数为1
func ChunkMap[K comparable, V any](m map[K]V, size int, fn func(chunk map[K]V), concurrent ...int) {
	if len(concurrent) == 0 {
		concurrent = []int{1}
	}
	if size <= 0 {
		return
	}
	if len(m) == 0 {
		return
	}
	length := len(m)

	wg := sync.WaitGroup{}
	ch := make(chan struct{}, concurrent[0])

	chunk := make(map[K]V)
	i := 0
	for key, value := range m {
		chunk[key] = value
		i++
		if i%size == 0 || i == length {
			wg.Add(1)
			ch <- struct{}{}
			go func(chunk map[K]V) {
				defer func() {
					wg.Done()
					<-ch
				}()
				fn(chunk)
			}(chunk)
			chunk = make(map[K]V)
		}
	}
	wg.Wait()
}

// SliceToMap 以slice中的item的某个字段值为key,item为value,转换为map
// 如果slice中有相同的key,则覆盖
func SliceToMap[T any, K comparable](s []T, fn func(index int, item T) (key K, value T)) map[K]T {
	m := make(map[K]T)
	for i, item := range s {
		key, value := fn(i, item)
		m[key] = value
	}
	return m
}

// SliceItemToSlice 将slice中的item转换为另一个slice
func SliceItemToSlice[T any, V any](s []T, fn func(index int, item T) V) []V {
	result := make([]V, 0, len(s))
	for i, item := range s {
		result = append(result, fn(i, item))
	}
	return result
}

// FilterSlice 过滤slice中的item,如果没有传递fn,则过滤掉为nil的item
func FilterSlice[T any](s []T, fn ...func(index int, item T) bool) []T {
	result := make([]T, 0, len(s))
	for i, item := range s {
		if len(fn) == 0 && !IsNil(item) {
			result = append(result, item)
		} else {
			if fn[0](i, item) {
				result = append(result, item)
			}
		}
	}
	return result
}

// FilterRepeat 过滤掉slice中的重复元素
func FilterRepeat[T comparable](s []T) []T {
	m := make(map[T]struct{})
	result := make([]T, 0, len(s))
	for _, item := range s {
		if _, ok := m[item]; !ok {
			m[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
