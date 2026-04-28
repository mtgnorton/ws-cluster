package tracing

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mtgnorton/ws-cluster/clustermessage"
	"github.com/mtgnorton/ws-cluster/logger"
	"github.com/mtgnorton/ws-cluster/shared"
	"github.com/mtgnorton/ws-cluster/shared/kit"
)

const payloadSnippetLimit = 240
const forceLogInterval = 2 * time.Second

const (
	EventWSRecvFromUser         = "ws_recv_from_user"
	EventWSRecvFromServer       = "ws_recv_from_server"
	EventWSPublishEnqueue       = "ws_publish_enqueue"
	EventWSRedisXAddDone        = "ws_redis_xadd_done"
	EventWSRedisXAddFailed      = "ws_redis_xadd_failed"
	EventWSRedisXRead           = "ws_redis_xread"
	EventWSQueueDispatchDone    = "ws_queue_dispatch_done"
	EventWSDispatchToServerDone = "ws_dispatch_to_server_done"
	EventWSDispatchToUserDone   = "ws_dispatch_to_user_done"
	EventServerQueueDequeue     = "server_queue_dequeue"
	EventServerWSWriteDone      = "server_ws_write_done"
	EventClientQueueDequeue     = "client_queue_dequeue"
	EventClientWSWriteDone      = "client_ws_write_done"
	EventServerSendQueueFull    = "server_send_queue_full"
	EventClientSendQueueFull    = "client_send_queue_full"
	EventServerWSWriteFailed    = "server_ws_write_failed"
	EventClientWSWriteFailed    = "client_ws_write_failed"
)

type Node struct {
	ID string
	IP string
}

type Meta struct {
	Type     clustermessage.Type
	PID      string
	AffairID string
	Payload  any
}

type Event struct {
	Name       string
	Reason     string
	DurationMs int64
	Force      bool
	Warn       bool
	Fields     map[string]any
}

type eventLog struct {
	Msg            string              `json:"msg"`
	TraceID        string              `json:"trace_id"`
	ParentID       string              `json:"parent_id,omitempty"`
	Type           clustermessage.Type `json:"type,omitempty"`
	PID            string              `json:"pid,omitempty"`
	AffairID       string              `json:"affair_id,omitempty"`
	Event          string              `json:"event"`
	Reason         string              `json:"reason,omitempty"`
	NodeID         string              `json:"node_id,omitempty"`
	NodeIP         string              `json:"node_ip,omitempty"`
	AtMs           int64               `json:"at_ms"`
	DurationMs     int64               `json:"duration_ms,omitempty"`
	PayloadSize    int                 `json:"payload_size,omitempty"`
	PayloadSnippet string              `json:"payload_snippet,omitempty"`
	Fields         map[string]any      `json:"fields,omitempty"`
}

var forceLogLimiters sync.Map

func CurrentNode() Node {
	return Node{
		ID: strconv.FormatInt(shared.GetNodeID(), 10),
		IP: shared.GetInternalIP(),
	}
}

func NodeInfo(nodeID string, nodeIP string) Node {
	return Node{ID: nodeID, IP: nodeIP}
}

func MetaFromMessage(msg *clustermessage.AffairMsg) Meta {
	if msg == nil {
		return Meta{}
	}
	return Meta{
		Type:     msg.Type,
		PID:      clustermessage.MessagePID(msg),
		AffairID: msg.AffairID,
		Payload:  msg.Payload,
	}
}

func StartMessage(ctx context.Context, l logger.Logger, msg *clustermessage.AffairMsg, node Node, eventName string) {
	clustermessage.MaybeStartTrace(msg, "")
	RecordMessage(ctx, l, msg, node, Event{Name: eventName})
}

func RecordMessage(ctx context.Context, l logger.Logger, msg *clustermessage.AffairMsg, node Node, event Event) {
	if msg == nil {
		return
	}
	msg.Trace = RecordTrace(ctx, l, msg.Trace, MetaFromMessage(msg), node, event)
}

func RecordTrace(ctx context.Context, l logger.Logger, trace *clustermessage.Trace, meta Meta, node Node, event Event) *clustermessage.Trace {
	if event.Name == "" || l == nil {
		return trace
	}
	reason := event.Reason
	if reason == "" {
		reason = event.Name
	}
	sampled := trace != nil && trace.ID != "" && trace.Sampled
	if event.Force && !sampled {
		if !allowForceLog(meta, node, event.Name, reason) {
			return trace
		}
		traceMsg := &clustermessage.AffairMsg{Trace: trace}
		trace = clustermessage.ForceTrace(traceMsg)
	}
	if trace == nil || trace.ID == "" || !trace.Sampled {
		return trace
	}
	line := buildLogLine(trace, meta, node, event, reason)
	if event.Warn {
		l.Warn(ctx, line)
		return trace
	}
	l.Info(ctx, line)
	return trace
}

func ShouldRecord(trace *clustermessage.Trace, force bool) bool {
	return force || (trace != nil && trace.ID != "" && trace.Sampled)
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

func buildLogLine(trace *clustermessage.Trace, meta Meta, node Node, event Event, reason string) string {
	payloadSize, payloadSnippet := payloadSummary(meta.Payload)
	line := eventLog{
		Msg:            "message trace",
		TraceID:        trace.ID,
		ParentID:       trace.ParentID,
		Type:           meta.Type,
		PID:            meta.PID,
		AffairID:       meta.AffairID,
		Event:          event.Name,
		Reason:         reason,
		NodeID:         node.ID,
		NodeIP:         node.IP,
		AtMs:           time.Now().UnixMilli(),
		DurationMs:     event.DurationMs,
		PayloadSize:    payloadSize,
		PayloadSnippet: payloadSnippet,
		Fields:         event.Fields,
	}
	b, err := json.Marshal(line)
	if err != nil {
		return `{"msg":"message trace","trace_id":"` + trace.ID + `","event":"` + event.Name + `","reason":"marshal_failed"}`
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
	if len(runes) > payloadSnippetLimit {
		text = string(runes[:payloadSnippetLimit]) + "..."
	}
	return size, text
}

func allowForceLog(meta Meta, node Node, event string, reason string) bool {
	key := node.ID + "|" + node.IP + "|" + string(meta.Type) + "|" + meta.PID + "|" + event + "|" + reason
	value, _ := forceLogLimiters.LoadOrStore(key, &atomic.Int64{})
	return kit.AllowByInterval(value.(*atomic.Int64), forceLogInterval)
}
