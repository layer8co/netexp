// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

// Package rcu provides a reference counter
// for implementing the read-copy-update (rcu) pattern.
package rcu

import (
	"sync"
)

type Rcu[T any] struct {
	mu       sync.Mutex
	itemPool sync.Pool
	latest   *Item[T]
	free     func(T)
}

type Item[T any] struct {
	v    T
	refs int
}

// The free function is called on old items that have no readers left.
func NewRcu[T any](free func(T)) *Rcu[T] {
	return &Rcu[T]{
		free: free,
		itemPool: sync.Pool{
			New: func() any {
				return &Item[T]{}
			},
		},
	}
}

func (r *Rcu[T]) Update(v T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	prev := r.latest
	r.latest = r.itemPool.Get().(*Item[T])
	r.latest.v = v
	if prev != nil {
		r.gc(prev)
	}
}

// Don't forget `defer r.ReadDone(handle)`.
func (r *Rcu[T]) Read() (v T, handle *Item[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item := r.latest
	if item == nil {
		return v, nil
	}
	item.refs++
	return item.v, item
	// Returning a closure that does the job of `r.ReadDone(handle)`
	// would cause an allocation, because it would close over `r *Rcu`,
	// and in order for the internal structure of the closure
	// to not escape to heap, either Get() would have to be inlined,
	// or the compiler would have to be advanced enough to realize
	// the scope of the closure could be as or more limited
	// than the variables it closes over, and do something about it.
	// However, neither of these are the case,
	// and Get() isn't inlined even without the defer statement.
	//
	// Relevant issues:
	//   - https://github.com/golang/go/issues/17566
	//   - https://github.com/golang/go/issues/21536
	//   - https://github.com/golang/go/issues/43210
}

func (r *Rcu[T]) ReadDone(handle *Item[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	handle.refs--
	if handle.refs < 0 {
		panic("rcu: negative reference counter")
	}
	r.gc(handle)
}

func (r *Rcu[T]) gc(item *Item[T]) {
	if item == r.latest || item.refs > 0 {
		return
	}
	if r.free != nil {
		r.free(item.v)
	}
	r.itemPool.Put(item)
}
