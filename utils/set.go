package utils

import (
	"sync"
)

// ReadOnlySet is a read-only set.
type ReadOnlySet[T comparable] struct {
	set map[T]struct{}
}

// NewReadOnlySet creates a new read-only set with the given values.
func NewReadOnlySet[T comparable](keys ...T) *ReadOnlySet[T] {
	set := make(map[T]struct{}, len(keys))
	for _, v := range keys {
		set[v] = struct{}{}
	}
	return &ReadOnlySet[T]{set: set}
}

// Len returns the number of elements in the set.
func (s *ReadOnlySet[T]) Len() int {
	return len(s.set)
}

// Has returns true if the set contains the given value.
func (s *ReadOnlySet[T]) Has(value T) bool {
	_, ok := s.set[value]
	return ok
}

// Values returns all the values in the set.
func (s *ReadOnlySet[T]) Values() []T {
	a := make([]T, len(s.set))
	i := 0
	for v := range s.set {
		a[i] = v
		i++
	}
	return a
}

// Set is a set.
type Set[T comparable] struct {
	lock sync.RWMutex
	set  map[T]struct{}
}

// NewSet creates a new set with the given values.
func NewSet[T comparable](keys ...T) *Set[T] {
	set := make(map[T]struct{}, len(keys))
	for _, v := range keys {
		set[v] = struct{}{}
	}
	return &Set[T]{set: set}
}

// Len returns the number of elements in the set.
func (s *Set[T]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.set)
}

// Has returns true if the set contains the given value.
func (s *Set[T]) Has(key T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.set[key]
	return ok
}

// Add adds the given value to the set.
func (s *Set[T]) Add(key T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.set[key] = struct{}{}
}

// Remove removes the given value from the set.
func (s *Set[T]) Remove(key T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.set, key)
}

// Reset removes all the values from the set.
func (s *Set[T]) Reset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	clear(s.set)
}

// Values returns all the values in the set.
func (s *Set[T]) Values() []T {
	s.lock.RLock()
	defer s.lock.RUnlock()

	a := make([]T, len(s.set))
	i := 0
	for v := range s.set {
		a[i] = v
		i++
	}
	return a
}
