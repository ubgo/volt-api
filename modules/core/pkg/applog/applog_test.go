package applog

import (
	"context"
	"testing"

	"github.com/ubgo/logger"
	"go.opentelemetry.io/otel/trace"
)

// captureSink records every emitted Record so tests can assert on Ctx/Fields.
type captureSink struct{ recs []logger.Record }

func (s *captureSink) Emit(r *logger.Record) error {
	// Records are pooled — copy the parts we assert on before release.
	s.recs = append(s.recs, logger.Record{Ctx: r.Ctx, Message: r.Message, Fields: append([]logger.Field(nil), r.Fields...)})
	return nil
}
func (s *captureSink) Sync() error  { return nil }
func (s *captureSink) Close() error { return nil }

// TestFromCtxBindsCtxToRecords pins the TraceId regression: FromCtx emissions
// MUST carry the caller's ctx on the Record (that ctx is what the OTEL sink
// hands the SDK to fill the binary TraceId column). A FromCtx that only adds
// string trace_id fields and emits with context.Background() silently breaks
// HyperDX's log→trace linking.
func TestFromCtxBindsCtxToRecords(t *testing.T) {
	sink := &captureSink{}
	prev := global
	global = logger.New(logger.WithSink(sink), logger.WithLevel(logger.LevelDebug))
	t.Cleanup(func() { global = prev })

	tid := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	sid := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: tid, SpanID: sid, TraceFlags: trace.FlagsSampled})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	FromCtx(ctx).Error("boom")

	if len(sink.recs) != 1 {
		t.Fatalf("emitted %d records, want 1", len(sink.recs))
	}
	got := trace.SpanContextFromContext(sink.recs[0].Ctx)
	if !got.IsValid() || got.TraceID() != tid {
		t.Fatalf("record.Ctx lost the span: valid=%v traceID=%s, want %s", got.IsValid(), got.TraceID(), tid)
	}
}

// TestFromCtxStringFieldsPresent — console/file sinks have no ctx extraction,
// so the string trace_id/span_id fields must still ride along when the ctx
// carries a valid span.
func TestFromCtxStringFieldsPresent(t *testing.T) {
	sink := &captureSink{}
	prev := global
	global = logger.New(logger.WithSink(sink), logger.WithLevel(logger.LevelDebug))
	t.Cleanup(func() { global = prev })

	tid := trace.TraceID{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9}
	sid := trace.SpanID{9, 9, 9, 9, 9, 9, 9, 9}
	sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: tid, SpanID: sid, TraceFlags: trace.FlagsSampled})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	FromCtx(ctx).With(logger.String("k", "v")).Info("hello")

	if len(sink.recs) != 1 {
		t.Fatalf("emitted %d records, want 1", len(sink.recs))
	}
	found := map[string]bool{}
	for _, f := range sink.recs[0].Fields {
		found[f.Key] = true
	}
	for _, want := range []string{"trace_id", "span_id", "k"} {
		if !found[want] {
			t.Errorf("field %q missing from record (got %v)", want, found)
		}
	}
}

// TestFromCtxNoSpan — a span-less ctx must still bind ctx (harmless for the
// OTEL sink) without inventing trace fields.
func TestFromCtxNoSpan(t *testing.T) {
	sink := &captureSink{}
	prev := global
	global = logger.New(logger.WithSink(sink), logger.WithLevel(logger.LevelDebug))
	t.Cleanup(func() { global = prev })

	FromCtx(context.Background()).Warn("plain")

	if len(sink.recs) != 1 {
		t.Fatalf("emitted %d records, want 1", len(sink.recs))
	}
	for _, f := range sink.recs[0].Fields {
		if f.Key == "trace_id" || f.Key == "span_id" {
			t.Errorf("unexpected %q field on span-less record", f.Key)
		}
	}
}
