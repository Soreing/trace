package trace

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/Soreing/motel"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type TraceCore struct {
	collector motel.SpanCollector
	exporters []sdktrace.SpanExporter
	rand      Random
}

// NewTraceCore creates a service that manages dispatching spans to exporters
// and provides utility functions for working with traces.
func NewTraceCore(
	exporters []sdktrace.SpanExporter,
	opts ...Option,
) (*TraceCore, error) {
	cfg, err := newConfiguration(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to configure: %w", err)
	}

	sc := motel.NewSpanCollector(exporters, cfg.batchTime, cfg.batchCount)
	return &TraceCore{
		collector: sc,
		exporters: exporters,
		rand:      cfg.rand,
	}, nil
}

// CreateResource creates an open telemetry resource with a name.
func (trc *TraceCore) CreateResource(
	ctx context.Context,
	serviceName string,
) (*resource.Resource, error) {
	attrib := semconv.ServiceNameKey.String(serviceName)
	return resource.New(ctx, resource.WithAttributes(attrib))
}

// CreateSpanId creates new [8]byte span id.
func (trc *TraceCore) CreateSpanId() (sid [8]byte) {
	trc.rand.Fill(sid[:])
	return
}

// CreateTraceId creates new [16]byte trace id.
func (trc *TraceCore) CreateTraceId() (tid [16]byte) {
	trc.rand.Fill(tid[:])
	return
}

// DispatchSpan submits a span to be dispatched by the exporters.
func (trc *TraceCore) DispatchSpan(span motel.Span) {
	trc.collector.Feed(span)
}

// Close closes the trace service and dispatches remaining spans.
func (trc *TraceCore) Close() {
	trc.collector.Close()
}

// CreateSpanId creates new [8]byte span id using entropy from crypto/rand.
func CreateSpanId() (sid [8]byte, err error) {
	_, err = crand.Read(sid[:])
	return
}

// CreateTraceId creates new [16]byte trace id using entropy from crypto/rand.
func CreateTraceId() (tid [16]byte, err error) {
	_, err = crand.Read(tid[:])
	return
}

// EncodeTraceparent creates a w3c traceparent header from the given version,
// trace id, parent id and flag bytes and byte arrays.
func EncodeTraceparent(ver byte, tid [16]byte, pid [8]byte, flg byte) string {
	header := make([]byte, 55)
	header[2], header[35], header[52] = '-', '-', '-'
	hex.Encode(header[:2], []byte{ver})
	hex.Encode(header[3:35], tid[:])
	hex.Encode(header[36:52], pid[:])
	hex.Encode(header[53:], []byte{flg})
	return string(header)
}

// DecodeTraceparent parses and validates a w3c traceparent header and returns
// the version, trace id, parent id and flag as bytes and byte arrays.
func DecodeTraceparent(
	header string,
) (ver byte, tid [16]byte, pid [8]byte, flg byte, err error) {
	var d1, d2, c uint8
	var val int

	if len(header) != 55 {
		err = fmt.Errorf("invalid length")
		return
	}

	if header[2] != '-' || header[35] != '-' || header[52] != '-' {
		err = fmt.Errorf("invalid format")
		return
	}

	// version
	switch header[:2] {
	case "00":
		ver = 0
	default:
		err = fmt.Errorf("invalid version")
		return
	}

	// flag
	switch header[53:] {
	case "00":
		flg = 0
	case "01":
		flg = 1
	default:
		err = fmt.Errorf("invalid flag")
		return
	}

	// trace id
	val = 0
	for i := 0; i < 16; i++ {
		c = header[3+(i<<1)]
		if d1 = c - '0'; d1 > 9 {
			if d1 = c - 'W'; d1 > 15 || d1 < 10 {
				err = fmt.Errorf("invalid trace id")
				return
			}
		}
		c = header[4+(i<<1)]
		if d2 = c - '0'; d2 > 9 {
			if d2 = c - 'W'; d2 > 15 || d2 < 10 {
				err = fmt.Errorf("invalid trace id")
				return
			}
		}
		tid[i] = (d1 << 4) + d2
		val += int(tid[i])
	}
	if val == 0 {
		err = fmt.Errorf("invalid trace id")
		return
	}

	// parent id
	val = 0
	for i := 0; i < 8; i++ {
		c = header[36+(i<<1)]
		if d1 = c - '0'; d1 > 9 {
			if d1 = c - 'W'; d1 > 15 || d1 < 10 {
				err = fmt.Errorf("invalid parent id")
				return
			}
		}
		c = header[37+(i<<1)]
		if d2 = c - '0'; d2 > 9 {
			if d2 = c - 'W'; d2 > 15 || d2 < 10 {
				err = fmt.Errorf("invalid parent id")
				return
			}
		}
		pid[i] = (d1 << 4) + d2
		val += int(pid[i])
	}
	if val == 0 {
		err = fmt.Errorf("invalid parent id")
		return
	}

	return
}
