package tracing

import "time"

func ExampleStart() {
	finish := Start()
	span := NewSpan("short")
	span.end = span.start.Add(time.Nanosecond * 12_300)
	span = NewSpan("long")
	span.end = span.start.Add(time.Microsecond * 23_400)
	finish()
	// Output:
	// Traced 2 span(s):
	// short 12us
	// long 23ms
}
