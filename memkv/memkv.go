package memkv

import (
	"sync"

	"github.com/wafer-bw/go-toolbox/memkv/internal/underlying"
)

// Store is a generic in-memory key-value store.
type Store[K comparable, V any] struct {
	mu       *sync.RWMutex
	capacity int
	data     *underlying.Data[K, V]
}

// New creates a new instance of [Store] with the provided capacity.
//
//   - A capacity of zero means the store has no capacity limit.
//   - If the capacity is less than 0, it will be set to 0.
func New[K comparable, V any](capacity int) *Store[K, V] {
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

// Set the provided key-value pair in the store.
func (s Store[K, V]) Set(key K, val V) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data.Items[key]; !ok && s.capacity > 0 && len(s.data.Items) >= s.capacity {
		return &AtCapacityError{}
	}

	s.data.Items[key] = underlying.Item[K, V]{Value: val}

	return nil
}

// Get the value associated with the provided key from the store if it exists.
func (s Store[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.data.Items[key]

	return item.Value, ok
}

// Delete provided keys from the store.
func (s Store[K, V]) Delete(keys ...K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		delete(s.data.Items, key)
	}
}

// Flush the cache, deleting all keys.
func (s Store[K, V]) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	clear(s.data.Items)
}

// Len returns the number of items currently in the store.
func (s Store[K, V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data.Items)
}

// Items returns a map of all items currently in the store.
func (s Store[K, V]) Items() map[K]V {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make(map[K]V, len(s.data.Items))
	for key, item := range s.data.Items {
		items[key] = item.Value
	}

	return items
}

// Keys returns a slice of all keys currently in the store.
func (s Store[K, V]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]K, 0, len(s.data.Items))
	for key := range s.data.Items {
		keys = append(keys, key)
	}

	return keys
}

// Values returns a slice of all values currently in the store.
func (s Store[K, V]) Values() []V {
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
