package memory

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/unistack-org/micro/v3/tracer"
	"github.com/unistack-org/micro/v3/util/ring"
)

type Tracer struct {
	opts tracer.Options

	// ring buffer of traces
	buffer *ring.Buffer
}

func (t *Tracer) Read(opts ...tracer.ReadOption) ([]*tracer.Span, error) {
	var options tracer.ReadOptions
	for _, o := range opts {
		o(&options)
	}

	sp := t.buffer.Get(t.buffer.Size())

	spans := make([]*tracer.Span, 0, len(sp))

	for _, span := range sp {
		val := span.Value.(*tracer.Span)
		// skip if trace id is specified and doesn't match
		if len(options.Trace) > 0 && val.Trace != options.Trace {
			continue
		}
		spans = append(spans, val)
	}

	return spans, nil
}

func (t *Tracer) Start(ctx context.Context, name string) (context.Context, *tracer.Span) {
	span := &tracer.Span{
		Name:     name,
		Trace:    uuid.New().String(),
		Id:       uuid.New().String(),
		Started:  time.Now(),
		Metadata: make(map[string]string),
	}

	// return span if no context
	if ctx == nil {
		return tracer.NewContext(context.Background(), span.Trace, span.Id), span
	}
	traceID, parentSpanID, ok := tracer.FromContext(ctx)
	// If the trace can not be found in the header,
	// that means this is where the trace is created.
	if !ok {
		return tracer.NewContext(ctx, span.Trace, span.Id), span
	}

	// set trace id
	span.Trace = traceID
	// set parent
	span.Parent = parentSpanID

	// return the span
	return tracer.NewContext(ctx, span.Trace, span.Id), span
}

func (t *Tracer) Finish(s *tracer.Span) error {
	// set finished time
	s.Duration = time.Since(s.Started)
	// save the span
	t.buffer.Put(s)

	return nil
}

func NewTracer(opts ...tracer.Option) tracer.Tracer {
	return &Tracer{
		opts: tracer.NewOptions(opts...),
		// the last 256 requests
		buffer: ring.New(256),
	}
}
