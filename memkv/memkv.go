package memkv

import (
	"sync"

	"github.com/wafer-bw/go-toolbox/memkv/internal/underlying"
)

type Store[K comparable, V any] struct { // TODO: docstring.
	mu       *sync.RWMutex
	capacity int
	data     *underlying.Data[K, V]
}

func New[K comparable, V any](capacity int) *Store[K, V] { // TODO: docstring.
	if capacity < 0 {
		capacity = 0
	}

	return &Store[K, V]{
		mu:       &sync.RWMutex{},
		capacity: capacity,
		data: &underlying.Data[K, V]{
			Items: make(map[K]underlying.Item[K, V], capacity),
		},
	}
}

func (s Store[K, V]) Set(key K, val V) error { // TODO: docstring.
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data.Items[key]; !ok && s.capacity > 0 && len(s.data.Items) >= s.capacity {
		return &AtCapacityError{}
	}

	s.data.Items[key] = underlying.Item[K, V]{Value: val}

	return nil
}

func (s Store[K, V]) Get(key K) (V, bool) { // TODO: docstring.
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.data.Items[key]

	return item.Value, ok
}

func (s Store[K, V]) Delete(keys ...K) { // TODO: docstring.
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		delete(s.data.Items, key)
	}
}

func (s Store[K, V]) Flush() { // TODO: docstring.
	s.mu.Lock()
	defer s.mu.Unlock()

	clear(s.data.Items)
}

func (s Store[K, V]) Len() int { // TODO: docstring.
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data.Items)
}

func (s Store[K, V]) Items() map[K]V { // TODO: docstring.
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make(map[K]V, len(s.data.Items))
	for key, item := range s.data.Items {
		items[key] = item.Value
	}

	return items
}

func (s Store[K, V]) Keys() []K { // TODO: docstring.
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]K, 0, len(s.data.Items))
	for key := range s.data.Items {
		keys = append(keys, key)
	}

	return keys
}

func (s Store[K, V]) Values() []V { // TODO: docstring.
	s.mu.RLock()
	defer s.mu.RUnlock()

	values := make([]V, 0, len(s.data.Items))
	for _, item := range s.data.Items {
		values = append(values, item.Value)
	}

	return values
}

// AtCapcityError occurs when the [Store] is at capacity and new items cannot be
// added.
type AtCapacityError struct{}

func (e *AtCapacityError) Error() string {
	return "store is at capacity"
}
