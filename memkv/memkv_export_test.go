package memkv

import "github.com/wafer-bw/go-toolbox/memkv/internal/underlying"

// export for testing.
func (s *Store[K, V]) Capacity() int {
	return s.capacity
}

// export for testing.
func (s *Store[K, V]) Data() (*underlying.Data[K, V], func()) {
	s.mu.Lock()
	return s.data, s.mu.Unlock
}
