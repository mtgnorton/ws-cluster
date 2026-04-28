package clustermessage

import (
	"crypto/rand"
	"encoding/hex"
	"hash/fnv"
	"strconv"
	"time"
)

type Trace struct {
	ID       string `json:"id,omitempty"`
	ParentID string `json:"parent_id,omitempty"`
	Sampled  bool   `json:"sampled,omitempty"`
	StartMs  int64  `json:"start_ms,omitempty"`
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

func shouldSample(traceID string) bool {
	if traceID == "" {
		return false
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(traceID))
	return h.Sum32()%1000 == 0
}
