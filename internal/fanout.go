package internal

import "github.com/samber/lo"

// Run three functions which return results or errors in parallel. If any return an error then
// return that error (starting with the first function). If none error then return their results.
func FanOut3[A any, B any, C any](
	a func() (A, error),
	b func() (B, error),
	c func() (C, error),
) (A, B, C, error) {
	ch1 := lo.Async2(a)
	ch2 := lo.Async2(b)
	ch3 := lo.Async2(c)

	var zero1 A
	var zero2 B
	var zero3 C

	r1, err := (<-ch1).Unpack()
	if err != nil {
		return zero1, zero2, zero3, err
	}
	r2, err := (<-ch2).Unpack()
	if err != nil {
		return zero1, zero2, zero3, err
	}
	r3, err := (<-ch3).Unpack()
	if err != nil {
		return zero1, zero2, zero3, err
	}

	return r1, r2, r3, nil
}
