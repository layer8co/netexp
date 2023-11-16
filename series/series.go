package series

type Series[T any] struct {
	Samples []T
	Keep int
}

func New[T any](keep int) *Series[T] {
	return &Series[T]{
		Keep: keep,
	}
}

func (t *Series[T]) Record(sample T) {
	t.Samples = append(t.Samples, sample)
	if t.Length() > t.Keep {
		t.Samples = t.Samples[t.Length() - t.Keep :]
	}
}

func (t *Series[T]) Length() int {
	return len(t.Samples)
}

func (t *Series[T]) Last(n int) T {
	return t.Samples[t.Length() - n]
}

func (t *Series[T]) Head(n int) []T {
	return t.Samples[t.Length() - n :]
}

func Map[A, B any](src *Series[A], fn func(A) B) *Series[B] {
	var b_samples []B
	for _, sample := range src.Samples {
		b_samples = append(b_samples, fn(sample))
	}
	return &Series[B]{
		Samples: b_samples,
		Keep: src.Keep,
	}
}
