// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

package rcu

import "sync"

type BufferRcu struct {
	rcu  *Rcu[*[]byte]
	pool sync.Pool
}

func NewBufferRcu() *BufferRcu {
	r := &BufferRcu{
		pool: sync.Pool{
			New: func() any {
				return &[]byte{}
			},
		},
	}
	r.rcu = NewRcu(r.poolPut)
	return r
}

func (r *BufferRcu) poolPut(b *[]byte) {
	*b = (*b)[:0]
	r.pool.Put(b)
}

func (r *BufferRcu) Update(fn func([]byte) ([]byte, error)) error {
	var err error
	buf := r.pool.Get().(*[]byte)
	*buf, err = fn(*buf)
	if err != nil {
		r.poolPut(buf)
		return err
	}
	r.rcu.Update(buf)
	return nil
}

func (r *BufferRcu) Read(fn func([]byte)) {
	buf, handle := r.rcu.Read()
	if handle == nil {
		fn(nil)
	} else {
		fn(*buf)
		r.rcu.ReadDone(handle)
	}
}
