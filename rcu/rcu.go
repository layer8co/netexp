// read, copy, update
package rcu

import (
	"slices"
	"sync"
)

// Represents a single unit of data that "RCU" Holds.
type Element[T any] struct {
	data     T
	mu       *sync.Mutex // guards the "refCount" down below.
	refCount int         // Read & Writes on "refCount" only happens under the mu lock.
}

// RCU is a structure that provides a safe way to Write and read
// data. All readers are guaranteed to access to the second latest
// buffer, Using its "Latest()" method.
type RCU[T any] struct {
	elements []Element[T]
	mu       sync.RWMutex
}

func New[T any]() *RCU[T] {
	return &RCU[T]{
		// 10 capacity guarantees that no reallocation occur, if and
		// only if the program doesn't append more than that. Which
		// is unlikely to happen if we configure a timeout deadline on
		// the HTTP server.
		elements: make([]Element[T], 0, 10),
		mu:       sync.RWMutex{},
	}
}

// Rotate adds a new instance of Element to the Elements slice and
// also removes unreferenced elements from the beginning of the slice.
func (rcu *RCU[T]) Rotate() {
	rcu.mu.Lock()
	defer rcu.mu.Unlock()

	newElem := Element[T]{
		refCount: 0,
		mu:       &sync.Mutex{},
	}

	rcu.elements = append(rcu.elements, newElem)

	if len(rcu.elements) <= 2 {
		return // So there is nothing to clean up.
	}

	// Only check up to last two (protect the last two: current and
	// previous elements). And do not waste your time if its lock
	// acquired.
	til := 0
	for i := 0; i < len(rcu.elements)-2; i++ {

		ok := rcu.elements[i].mu.TryLock()

		if !ok {
			break
		}

		if rcu.elements[i].refCount > 0 {
			rcu.elements[i].mu.Unlock()
			break // Stop if we hit a referenced element; We only remove consecutive unreferenced elements.
		}
		til++
		rcu.elements[i].mu.Unlock()
	}

	if til > 0 {
		rcu.elements = slices.Delete(rcu.elements, 0, til)
	}
}

type RefDecrementFunc func()

// returns the most recent valid element. The caller is reponsible for
// decrementing the refCount using the returned "RefDecrementFunc".
func (rcu *RCU[T]) Latest() (*T, RefDecrementFunc) {
	rcu.mu.RLock()

	if len(rcu.elements) >= 2 {
		index := len(rcu.elements) - 2

		elem := &rcu.elements[index]
		rcu.mu.RUnlock()

		elem.mu.Lock()
		elem.refCount++
		elem.mu.Unlock()

		return &elem.data, func() {
			elem.mu.Lock()
			elem.refCount--
			elem.mu.Unlock()
		}
	}

	rcu.mu.RUnlock()
	return nil, nil
}

// Assigns data to the last index of "elements" slice.  It doesn't
// need mutual exclution, because only one goroutine manipulates the
// rcu slice.
func (rcu *RCU[T]) Assign(data T) {
	l := len(rcu.elements)
	rcu.elements[l-1].data = data
}
