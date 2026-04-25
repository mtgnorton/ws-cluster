package clustermessage

import (
	"crypto/rand"
	"encoding/hex"
	"hash/fnv"
	"strconv"
	"strings"
	"time"
)

const maxTraceEvents = 16
const tracePayloadSnippetLimit = 240

const (
	TraceEventWSRecvFromUser         = "ws_recv_from_user"
	TraceEventWSRecvFromServer       = "ws_recv_from_server"
	TraceEventTraceMissingFromBiz    = "trace_missing_from_business"
	TraceEventWSPublishEnqueue       = "ws_publish_enqueue"
	TraceEventWSRedisXAddDone        = "ws_redis_xadd_done"
	TraceEventWSRedisXRead           = "ws_redis_xread"
	TraceEventWSDispatchToServerDone = "ws_dispatch_to_server_done"
	TraceEventWSDispatchToUserDone   = "ws_dispatch_to_user_done"
	TraceEventServerQueueDequeue     = "server_queue_dequeue"
	TraceEventServerWSWriteDone      = "server_ws_write_done"
	TraceEventClientQueueDequeue     = "client_queue_dequeue"
	TraceEventClientWSWriteDone      = "client_ws_write_done"
	TraceEventServerSendQueueFull    = "server_send_queue_full"
	TraceEventClientSendQueueFull    = "client_send_queue_full"
	TraceEventServerWSWriteFailed    = "server_ws_write_failed"
	TraceEventClientWSWriteFailed    = "client_ws_write_failed"
)

type Trace struct {
	ID       string       `json:"id,omitempty"`
	ParentID string       `json:"parent_id,omitempty"`
	Sampled  bool         `json:"sampled,omitempty"`
	StartMs  int64        `json:"start_ms,omitempty"`
	Events   []TraceEvent `json:"events,omitempty"`
}

type TraceEvent struct {
	Name     string `json:"name"`
	AtMs     int64  `json:"at_ms"`
	NodeID   string `json:"node_id,omitempty"`
	NodeIP   string `json:"node_ip,omitempty"`
	Duration int64  `json:"duration_ms,omitempty"`
}

type TraceLog struct {
	Msg            string       `json:"msg"`
	TraceID        string       `json:"trace_id"`
	ParentID       string       `json:"parent_id,omitempty"`
	Type           Type         `json:"type,omitempty"`
	PID            string       `json:"pid,omitempty"`
	AffairID       string       `json:"affair_id,omitempty"`
	Reason         string       `json:"reason,omitempty"`
	NodeID         string       `json:"node_id,omitempty"`
	NodeIP         string       `json:"node_ip,omitempty"`
	PayloadSize    int          `json:"payload_size,omitempty"`
	PayloadSnippet string       `json:"payload_snippet,omitempty"`
	Events         []TraceEvent `json:"events,omitempty"`
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

func (t *Trace) Clone() *Trace {
	if t == nil {
		return nil
	}
	cp := *t
	if len(t.Events) > 0 {
		cp.Events = append([]TraceEvent(nil), t.Events...)
	}
	return &cp
}

func AddTraceEvent(trace *Trace, name string, nodeID string, nodeIP string, durationMs int64, force bool) {
	if trace == nil || name == "" {
		return
	}
	if !trace.Sampled && !force {
		return
	}
	if len(trace.Events) >= maxTraceEvents {
		return
	}
	if durationMs < 0 {
		durationMs = 0
	}
	trace.Events = append(trace.Events, TraceEvent{
		Name:     name,
		AtMs:     time.Now().UnixMilli(),
		NodeID:   nodeID,
		NodeIP:   nodeIP,
		Duration: durationMs,
	})
}

func AddTraceEventToMessage(msg *AffairMsg, name string, nodeID string, nodeIP string, durationMs int64, force bool) {
	if msg == nil {
		return
	}
	AddTraceEvent(msg.Trace, name, nodeID, nodeIP, durationMs, force)
}

func DurationMs(d time.Duration) int64 {
	if d <= 0 {
		return 0
	}
	ms := d.Milliseconds()
	if ms == 0 {
		return 1
	}
	return ms
}

func ShouldLogTrace(trace *Trace, force bool) bool {
	return trace != nil && trace.ID != "" && (trace.Sampled || force)
}

func BuildTraceLogForMessage(msg *AffairMsg, nodeID string, nodeIP string, reason string) string {
	if msg == nil {
		return ""
	}
	return BuildTraceLogForPayload(msg.Trace, msg.Type, MessagePID(msg), msg.AffairID, nodeID, nodeIP, reason, msg.Payload)
}

func BuildTraceLogForPayload(trace *Trace, msgType Type, pid string, affairID string, nodeID string, nodeIP string, reason string, payload any) string {
	if trace == nil || trace.ID == "" {
		return ""
	}
	payloadSize, payloadSnippet := payloadSummary(payload)
	line := TraceLog{
		Msg:            "message trace",
		TraceID:        trace.ID,
		ParentID:       trace.ParentID,
		Type:           msgType,
		PID:            pid,
		AffairID:       affairID,
		Reason:         reason,
		NodeID:         nodeID,
		NodeIP:         nodeIP,
		PayloadSize:    payloadSize,
		PayloadSnippet: payloadSnippet,
		Events:         trace.Events,
	}
	b, err := json.Marshal(line)
	if err != nil {
		return `{"msg":"message trace","trace_id":"` + trace.ID + `","reason":"marshal_failed"}`
	}
	return string(b)
}

func payloadSummary(payload any) (int, string) {
	if payload == nil {
		return 0, ""
	}
	var text string
	switch v := payload.(type) {
	case string:
		text = v
	case []byte:
		text = string(v)
	default:
		b, err := json.Marshal(payload)
		if err != nil {
			text = "<payload_marshal_failed>"
			break
		}
		text = string(b)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return 0, ""
	}
	size := len([]byte(text))
	runes := []rune(text)
	if len(runes) > tracePayloadSnippetLimit {
		text = string(runes[:tracePayloadSnippetLimit]) + "..."
	}
	return size, text
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
