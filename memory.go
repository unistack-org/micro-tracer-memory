package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/micro/go-micro/debug/trace"
	"github.com/micro/go-micro/util/ring"
)

type Tracer struct {
	opts trace.Options

	// ring buffer of traces
	buffer *ring.Buffer
}

func (t *Tracer) Read(opts ...trace.ReadOption) ([]*trace.Span, error) {
	var options trace.ReadOptions
	for _, o := range opts {
		o(&options)
	}

	sp := t.buffer.Get(t.buffer.Size())

	var spans []*trace.Span

	for _, span := range sp {
		val := span.Value.(*trace.Span)
		// skip if trace id is specified and doesn't match
		if len(options.Trace) > 0 && val.Trace != options.Trace {
			continue
		}
		spans = append(spans, val)
	}

	return spans, nil
}

func (t *Tracer) Start(ctx context.Context, name string) (context.Context, *trace.Span) {
	span := &trace.Span{
		Name:     name,
		Trace:    uuid.New().String(),
		Id:       uuid.New().String(),
		Started:  time.Now(),
		Metadata: make(map[string]string),
	}

	// return span if no context
	if ctx == nil {
		return context.Background(), span
	}

	s, ok := trace.FromContext(ctx)
	if !ok {
		return ctx, span
	}

	// set trace id
	span.Trace = s.Trace
	// set parent
	span.Parent = s.Id

	// return the sapn
	return ctx, span
}

func (t *Tracer) Finish(s *trace.Span) error {
	// set finished time
	s.Duration = time.Since(s.Started)

	// save the span
	t.buffer.Put(s)

	return nil
}

func NewTracer(opts ...trace.Option) trace.Tracer {
	var options trace.Options
	for _, o := range opts {
		o(&options)
	}

	return &Tracer{
		opts: options,
		// the last 64 requests
		buffer: ring.New(64),
	}
}
