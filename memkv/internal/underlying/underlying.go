// Package underlying provides the underlying data structures for the key-value
// store.
//
// This makes it easy to extend out the data structures in the future and
// enables easy testing of the store's underlying data via memkv.Store.Data
// (see memkv_export_test.go).
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
