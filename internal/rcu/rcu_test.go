// Copyright 2023 the netexp authors.
// SPDX-License-Identifier: MIT

package rcu_test

import (
	"testing"

	"github.com/layer8co/netexp/internal/rcu"
	"github.com/stretchr/testify/assert"
)

func TestRcu_Simple(t *testing.T) {

	ptrMap := make(map[*int]bool)

	r := rcu.NewRcu(func(ptr *int) {
		delete(ptrMap, ptr)
	})

	v, handle := r.Read()
	assertEqual(t, v, nil)
	if handle != nil {
		t.Fatalf("Rcu.Read() on empty refpool returns non-nil function")
	}

	A := new(int)
	ptrMap[A] = true
	r.Update(A)

	B := new(int)
	ptrMap[B] = true
	r.Update(B)

	C := new(int)
	ptrMap[C] = true
	r.Update(C)

	D := new(int)
	ptrMap[D] = true
	r.Update(D)

	assertEqual(t, len(ptrMap), 1)
	assert.Contains(t, ptrMap, D)
}

func TestRcu_Full(t *testing.T) {

	ptrMap := make(map[*int]bool)

	r := rcu.NewRcu(func(ptr *int) {
		delete(ptrMap, ptr)
	})

	// Put A
	A := new(int)
	ptrMap[A] = true
	r.Update(A)

	// Request latest, assert that it's A
	gotA, handleA := r.Read()
	assertEqual(t, gotA, A)
	assertEqual(t, len(ptrMap), 1)
	assert.Contains(t, ptrMap, A)

	// Put B
	B := new(int)
	ptrMap[B] = true
	r.Update(B)

	// Request latest, assert that it's B
	gotB, handleB := r.Read()
	assertEqual(t, gotB, B)
	assertEqual(t, len(ptrMap), 2)
	assert.Contains(t, ptrMap, A, B)

	// Request latest again, assert that it's B
	gotB2, handleB2 := r.Read()
	assertEqual(t, gotB2, B)
	assertEqual(t, len(ptrMap), 2)
	assert.Contains(t, ptrMap, A, B)

	// Put C
	C := new(int)
	ptrMap[C] = true
	r.Update(C)

	// Finish request A
	r.ReadDone(handleA)
	assertEqual(t, len(ptrMap), 2)
	assert.Contains(t, ptrMap, B, C)

	// Finish request B
	r.ReadDone(handleB)
	assertEqual(t, len(ptrMap), 2)
	assert.Contains(t, ptrMap, B, C)

	// Put D
	D := new(int)
	ptrMap[D] = true
	r.Update(D)

	// Finish request B2.
	// Since C was not requested, it gets removed.
	r.ReadDone(handleB2)
	assertEqual(t, len(ptrMap), 1)
	assert.Contains(t, ptrMap, D)

	// Request latest, assert that it's D
	gotD, _ := r.Read()
	assertEqual(t, gotD, D)
	assertEqual(t, len(ptrMap), 1)
	assert.Contains(t, ptrMap, D)
}

func TestRcu_NoAlloc(t *testing.T) {
	const wantAllocs = 0
	r := rcu.NewRcu[*int](nil)
	p := new(int)
	allocs := testing.AllocsPerRun(100, func() {
		r.Update(p)
		_, handle := r.Read()
		r.ReadDone(handle)
	})
	assertEqual(t, allocs, wantAllocs)
}

// [2025-12-25] 140 ns/op, 0 allocs/op on an 8-core 10th-gen Intel.
func BenchmarkRcu(b *testing.B) {
	r := rcu.NewRcu[*int](nil)
	p := new(int)
	for b.Loop() {
		r.Update(p)
		_, handle := r.Read()
		r.ReadDone(handle)
	}
}

func assertEqual[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}
