package tracing

import (
	"fmt"
	"time"
)

type Span struct {
	name  string
	start time.Time
	end   time.Time
}

func NewSpan(name string) *Span {
	start := time.Now()
	span := &Span{
		name:  name,
		start: start,
		end:   start,
	}
	spans = append(spans, span)
	return span
}

func (span *Span) End() {
	span.end = time.Now()
}

func (span *Span) Elapsed() time.Duration {
	return span.end.Sub(span.start)
}

var spans []*Span

func Start() func() {
	spans = []*Span{}
	return func() {
		fmt.Printf("Traced %v span(s):\n", len(spans))
		for _, span := range spans {
			elapsed := span.Elapsed()
			millis := elapsed.Milliseconds()
			var human string
			if millis <= 2 {
				human = fmt.Sprintf("%dus", elapsed.Microseconds())
			} else {
				human = fmt.Sprintf("%dms", millis)
			}
			fmt.Printf("%s %s\n", span.name, human)
		}
	}
}
