package utils

import (
	"sync"
)

// ReadOnlySet is a read-only set.
type ReadOnlySet[T comparable] struct {
	values map[T]struct{}
}

// NewReadOnlySet creates a new read-only set with the given values.
func NewReadOnlySet[T comparable](values ...T) *ReadOnlySet[T] {
	if len(values) == 0 {
		return &ReadOnlySet[T]{}
	}
	set := make(map[T]struct{}, len(values))
	for _, v := range values {
		set[v] = struct{}{}
	}
	return &ReadOnlySet[T]{values: set}
}

// Len returns the number of elements in the set.
func (s *ReadOnlySet[T]) Len() int {
	return len(s.values)
}

// Has returns true if the set contains the given value.
func (s *ReadOnlySet[T]) Has(value T) bool {
	_, ok := s.values[value]
	return ok
}

// Values returns all the values in the set.
func (s *ReadOnlySet[T]) Values() []T {
	a := make([]T, len(s.values))
	i := 0
	for v := range s.values {
		a[i] = v
		i++
	}
	return a
}

// Set is a set.
type Set[T comparable] struct {
	lock   sync.RWMutex
	values map[T]struct{}
}

// NewSet creates a new set with the given values.
func NewSet[T comparable](values ...T) *Set[T] {
	set := make(map[T]struct{}, len(values))
	for _, v := range values {
		set[v] = struct{}{}
	}
	return &Set[T]{values: set}
}

// Len returns the number of elements in the set.
func (s *Set[T]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.values)
}

// Has returns true if the set contains the given value.
func (s *Set[T]) Has(value T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.values[value]
	return ok
}

// Add adds the given value to the set.
func (s *Set[T]) Add(value T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.values[value] = struct{}{}
}

// Remove removes the given value from the set.
func (s *Set[T]) Remove(value T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.values, value)
}

// Reset removes all the values from the set.
func (s *Set[T]) Reset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	clear(s.values)
}

// Values returns all the values in the set.
func (s *Set[T]) Values() []T {
	s.lock.RLock()
	defer s.lock.RUnlock()

	a := make([]T, len(s.values))
	i := 0
	for v := range s.values {
		a[i] = v
		i++
	}
	return a
}
