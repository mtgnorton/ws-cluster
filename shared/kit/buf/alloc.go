package buf

// https://github.com/xtaci/smux/blob/master/alloc.go
import (
	"errors"
	"sync"
)

var (
	DefaultAllocator *Allocator = NewAllocator()
	debruijinPos                = [...]byte{0, 9, 1, 10, 13, 21, 2, 29, 11, 14, 16, 18, 22, 25, 3, 30, 8, 12, 20, 28, 15, 17, 24, 7, 19, 27, 23, 6, 26, 5, 4, 31}
)

// Allocator 是[]byte的分配器，
// 优点: 避免重复创建[]byte,降低内存占用,减少GC压力
// 场景: io读写, 其他需要频繁创建[]byte的场景
// 缺点: 需要手动释放, 需要手动管理内存
//
// 什么是内存碎片
//     golang底层的内存管理最小单位是mspan,根据管理的对象大小从8 bytes到32KB不等,共67中大小,mspan的空间来自于arena的单个或多个页,每页的大小是8kb,即8192 bytes,
//     假设我们先创建了一个变量a,占用5kb,又创建了一个变量b,占用2kb,又创建了一个变量c,占用1kb,然后释放了变量b,此时变量a和变量c之间就产生了1kb的内存碎片
// 什么是内存对齐
//	  CPU通常以字(word)为单位读取内存, 32位CPU的内存访问粒度是4字节,64位CPU的内存访问粒度是8字节,
//     内存对齐的目的是为了提高内存访问效率, 例如: 32位CPU访问一个4字节的数据, 64位CPU访问一个8字节的数据, 如果不对齐, 则需要两次内存访问, 对齐后只需要一次内存访问
// 	未对齐访问
// 		CPU可能需要：1. 读取第一个内存块 2. 读取第二个内存块 3. 合并数据 4. 提取需要的部分
// 对齐访问
// 		CPU只需：1. 一次读取操作 2. 直接使用数据

// 为什么设置固定的等级?
//  1. 降低内存碎片,每个大小都不同，导致内存碎片化严重
//  2. 为什么容量都为为2^n?
//     a. 不容易产生碎片
//     b. 内存对齐, 2^n是内存对齐的单位
//     c. 内存访问的粒度, 例如: 32位CPU访问一个4字节的数据, 64位CPU访问一个8字节的数据
type Allocator struct {
	buffers []sync.Pool
}

// NewAllocator 为小于65536字节的[]byte初始化分配器，
// 保证空间分配的浪费(内存碎片)不超过50%。
func NewAllocator() *Allocator {
	alloc := new(Allocator)
	alloc.buffers = make([]sync.Pool, 17) // 1B -> 64K
	for k := range alloc.buffers {
		i := k
		alloc.buffers[k].New = func() interface{} {
			return make([]byte, 1<<uint32(i))
		}
	}
	return alloc
}

// Get 从池中获取一个容量最合适的[]byte
func (alloc *Allocator) Get(size int) []byte {
	if size <= 0 || size > 65536 {
		return nil
	}

	bits := msb(size)
	if size == 1<<bits {
		return alloc.buffers[bits].Get().([]byte)[:size]
	} else {
		return alloc.buffers[bits+1].Get().([]byte)[:size]
	}
}

// Put 将[]byte返回到池中以供将来使用，
// 其容量必须恰好为2^n
func (alloc *Allocator) Put(buf []byte) error {
	bits := msb(cap(buf))
	if cap(buf) == 0 || cap(buf) > 65536 || cap(buf) != 1<<bits {
		return errors.New("allocator Put() incorrect buffer size")
	}
	alloc.buffers[bits].Put(buf)
	return nil
}

// msb 返回最高有效位的位置
// msb 使用位运算计算一个整数的最高有效位(Most Significant Bit)位置
// 例如:
// 输入: 8 (二进制:1000) 返回: 3
// 输入: 7 (二进制:0111) 返回: 2
func msb(size int) byte {
	// 将输入转换为32位无符号整数
	v := uint32(size)

	// 通过一系列右移和或运算,将最高位1右边的所有位都设置为1
	// 例如输入8(1000),经过下面的运算后变成15(1111)
	v |= v >> 1  // 1000 | 0100 = 1100
	v |= v >> 2  // 1100 | 0011 = 1111
	v |= v >> 4  // 1111 | 0000 = 1111
	v |= v >> 8  // 1111 | 0000 = 1111
	v |= v >> 16 // 1111 | 0000 = 1111

	// 使用De Bruijn序列快速查找最高位1的位置
	// 0x07C4ACDD是一个精心设计的魔数,可以配合debruijinPos表实现O(1)时间复杂度的查找
	return debruijinPos[(v*0x07C4ACDD)>>27]
}
