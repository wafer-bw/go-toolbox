// Package underlying provides the underlying data structures for the key-value
// store.
package underlying

// Item is a wrapper around the instances of data to be stored allowing for
// extensions in the future.
type Item[K comparable, V any] struct {
	Value V
}

// Data is a wrapper around any data types used to store data in the store
// allowing for extensions in the future.
type Data[K comparable, V any] struct {
	Items map[K]Item[K, V]
}
