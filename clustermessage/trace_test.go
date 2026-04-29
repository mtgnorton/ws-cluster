// 模块职责说明：验证消息链路 Trace 创建、保留和采样配置行为。
package clustermessage

import "testing"

func TestMaybeStartTraceKeepsUnsampledTrace(t *testing.T) {
	msg := &AffairMsg{
		Trace: &Trace{ID: "trace-id", Sampled: false},
	}

	trace := MaybeStartTrace(msg, "")

	if trace == nil {
		t.Fatal("expected existing trace to be kept")
	}
	if msg.Trace == nil || msg.Trace.ID != "trace-id" {
		t.Fatalf("expected message trace to be preserved, got %+v", msg.Trace)
	}
}

func TestMaybeStartTraceKeepsSampledTrace(t *testing.T) {
	msg := &AffairMsg{
		Trace: &Trace{ID: "trace-id", Sampled: true},
	}

	trace := MaybeStartTrace(msg, "")

	if trace == nil {
		t.Fatal("expected sampled trace")
	}
	if msg.Trace == nil || msg.Trace.ID != "trace-id" {
		t.Fatalf("expected original trace to be kept, got %+v", msg.Trace)
	}
}

func TestMaybeStartTraceCreatesSampledChildTrace(t *testing.T) {
	msg := &AffairMsg{}

	trace := MaybeStartTrace(msg, "parent-id")

	if trace == nil {
		t.Fatal("expected child trace")
	}
	if !trace.Sampled {
		t.Fatal("expected child trace to be sampled")
	}
	if trace.ParentID != "parent-id" {
		t.Fatalf("expected parent-id, got %q", trace.ParentID)
	}
}

func TestForceTraceCreatesSampledTrace(t *testing.T) {
	msg := &AffairMsg{}

	trace := ForceTrace(msg)

	if trace == nil {
		t.Fatal("expected forced trace")
	}
	if msg.Trace == nil || msg.Trace.ID == "" {
		t.Fatalf("expected message trace to be populated, got %+v", msg.Trace)
	}
	if !trace.Sampled {
		t.Fatal("expected forced trace to be sampled")
	}
}

func TestShouldSampleUsesConfiguredDenominator(t *testing.T) {
	oldDenominator := traceSampleDenominator.Load()
	t.Cleanup(func() {
		SetTraceSampleDenominator(int(oldDenominator))
	})

	SetTraceSampleDenominator(1)
	if !shouldSample("trace-id") {
		t.Fatal("expected denominator 1 to sample every trace")
	}

	SetTraceSampleDenominator(0)
	if shouldSample("trace-id") {
		t.Fatal("expected denominator 0 to disable random sampling")
	}
}
