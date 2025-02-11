package kit

import (
	"errors"
	"sync"
	"time"
)

const (
	epoch          = int64(1700465775306)              // 设置起始时间(时间戳/毫秒)
	timestampBits  = uint(41)                          // 时间戳占用位数
	nodeIdBits     = uint(6)                           // 机器id所占位数
	sequenceBits   = uint(16)                          // 序列所占的位数
	timestampMax   = int64(-1 ^ (-1 << timestampBits)) // 时间戳最大值
	nodeIdMax      = int64(-1 ^ (-1 << nodeIdBits))    // 支持的最大机器id数量
	sequenceMask   = int64(-1 ^ (-1 << sequenceBits))  // 支持的最大序列id数量
	nodeIdShift    = sequenceBits                      // 机器id左移位数
	timestampShift = sequenceBits + nodeIdBits         // 时间戳左移位数
)

type UniqueNode struct {
	sync.Mutex
	nodeID    int64 // 机器ID
	sequence  int64 // 序列号
	timestamp int64 // 时间戳 ，毫秒
}

var DefaultUniqueNode = &UniqueNode{nodeID: 1}

func (s *UniqueNode) Generate() int64 {
	s.Lock()
	defer s.Unlock()
	now := time.Now().UnixMilli() // 转毫秒
	if s.timestamp == now {
		// 当同一时间戳（精度：毫秒）下多次生成id会增加序列号
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 如果当前序列超出12bit长度，则需要等待下一毫秒
			// 下一毫秒将使用sequence:0
			for now <= s.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// 不同时间戳（精度：毫秒）下直接使用序列号：0
		s.sequence = 0
	}
	t := now - epoch
	if t > timestampMax {
		return 0
	}
	s.timestamp = now
	r := t<<timestampShift | (s.nodeID << nodeIdShift) | (s.sequence)

	return r
}

func NewUniqueNode(nodeId int64) (*UniqueNode, error) {
	if nodeId < 0 || nodeId > nodeIdMax {
		return nil, errors.New("nodeID must be between 0 and 255")
	}
	return &UniqueNode{nodeID: nodeId}, nil
}

// GenerateUniqueId 生成唯一ID
func GenerateUniqueId(node ...uint) (int64, error) {
	var (
		u   *UniqueNode
		err error
	)
	if len(node) > 0 {
		u, err = NewUniqueNode(int64(node[0]))
		if err != nil {
			return 0, err
		}
	} else {
		u = DefaultUniqueNode
	}

	return u.Generate(), nil
}
