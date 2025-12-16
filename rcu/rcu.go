package rcu

import "sync/atomic"

type RCU[T any] struct {
	p atomic.Pointer[T]
}

func (r *RCU[T]) Store(t *T) {
	r.p.Store(t)
}

func (r *RCU[T]) Load() *T {
	return r.p.Load()
}
