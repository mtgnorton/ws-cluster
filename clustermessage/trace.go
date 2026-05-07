// 模块职责说明：负责消息链路 Trace 的创建、采样和上下文传递。
package clustermessage

import (
	"crypto/rand"
	"encoding/hex"
	"hash/fnv"
	"strconv"
	"sync/atomic"
	"time"
)

const defaultTraceSampleDenominator uint32 = 1000 // 保留原有约 1/1000 的入口采样比例。

var traceSampleDenominator atomic.Uint32

func init() {
	traceSampleDenominator.Store(defaultTraceSampleDenominator)
}

type Trace struct {
	ID       string `json:"id,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
	Sampled  bool   `json:"sampled,omitempty"`
	StartMs  int64  `json:"start_ms,omitempty"`
}

// SetTraceSampleDenominator 函数说明：
// 功能：设置入口随机 Trace 采样分母。
// 输入：denominator 为采样分母，1000 表示约 1/1000，小于等于 0 表示关闭随机采样。
// 输出：无。
// 关键约束：运行期可能被测试或初始化流程读取，使用 atomic 避免数据竞争。
// 为什么这样实现：采样逻辑保持在消息模块内，启动流程只负责把配置值注入进来。
func SetTraceSampleDenominator(denominator int) {
	if denominator <= 0 {
		traceSampleDenominator.Store(0)
		return
	}
	traceSampleDenominator.Store(uint32(denominator))
}

func newTrace(parentID string) *Trace {
	traceID := newTraceID()
	return &Trace{
		ID:       traceID,
		ParentID: parentID,
		Sampled:  shouldSample(traceID),
		StartMs:  time.Now().UnixMilli(),
	}
}

func EnsureTrace(msg *AffairMsg, parentID string) *Trace {
	if msg == nil {
		return nil
	}
	if msg.Trace == nil {
		msg.Trace = newTrace(parentID)
		return msg.Trace
	}
	if msg.Trace.ID == "" {
		msg.Trace.ID = newTraceID()
		msg.Trace.Sampled = shouldSample(msg.Trace.ID)
	}
	if msg.Trace.StartMs == 0 {
		msg.Trace.StartMs = time.Now().UnixMilli()
	}
	if msg.Trace.ParentID == "" && parentID != "" {
		msg.Trace.ParentID = parentID
	}
	return msg.Trace
}

// MaybeStartTrace 在链路入口按采样率创建 trace。
// 已有 trace 必须保留，不能因为 sampled=false 清掉上游上下文。
func MaybeStartTrace(msg *AffairMsg, parentID string) *Trace {
	if msg == nil {
		return nil
	}
	if msg.Trace != nil {
		return EnsureTrace(msg, parentID)
	}

	traceID := newTraceID()
	if parentID == "" && !shouldSample(traceID) {
		return nil
	}
	msg.Trace = &Trace{
		ID:       traceID,
		ParentID: parentID,
		Sampled:  true,
		StartMs:  time.Now().UnixMilli(),
	}
	return msg.Trace
}

// ForceTrace 用于慢请求、失败、队列满等诊断路径。
// 一旦强制进入 trace，后续链路也按 sampled=true 输出事件日志。
func ForceTrace(msg *AffairMsg) *Trace {
	trace := EnsureTrace(msg, "")
	if trace != nil {
		trace.Sampled = true
	}
	return trace
}

func (t *Trace) Clone() *Trace {
	if t == nil {
		return nil
	}
	cp := *t
	return &cp
}

func MessagePID(msg *AffairMsg) string {
	if msg == nil {
		return ""
	}
	if msg.To != nil && msg.To.PID != "" {
		return msg.To.PID
	}
	if msg.Source != nil {
		return msg.Source.PID
	}
	return ""
}

func newTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// shouldSample 函数说明：
// 功能：按配置的采样分母稳定判断当前 Trace 是否采样。
// 输入：traceID 为链路 ID。
// 输出：true 表示采样，false 表示不采样。
// 关键约束：相同 traceID 必须得到稳定结果，避免链路中途采样状态抖动。
// 为什么这样实现：FNV 哈希成本低、无随机状态，适合高频消息路径。
func shouldSample(traceID string) bool {
	if traceID == "" {
		return false
	}
	denominator := traceSampleDenominator.Load()
	if denominator == 0 {
		return false
	}
	if denominator == 1 {
		// 分母为 1 表示全量采样，避免 hash 计算浪费在最高频配置上。
		return true
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(traceID))
	return h.Sum32()%denominator == 0
}
