package memkv_test

import (
	"fmt"

	"github.com/wafer-bw/go-toolbox/memkv"
)

func ExampleStore_Set() {
	store := memkv.New[string, string](0)

	key, val := "key", "val"
	if err := store.Set(key, val); err != nil {
		fmt.Println(err)
	}

	// Output:
}

func ExampleStore_Get() {
	store := memkv.New[string, string](0)

	key, val := "key", "val"
	if err := store.Set(key, val); err != nil {
		return
	}

	v, ok := store.Get(key)
	fmt.Println(v, ok)

	// Output: val true
}

func ExampleStore_Delete() {
	store := memkv.New[string, string](0)

	key, val := "key", "val"
	if err := store.Set(key, val); err != nil {
		return
	}

	v, ok := store.Get(key)
	fmt.Println(v, ok)

	store.Delete(key)

	v, ok = store.Get(key)
	fmt.Println(v, ok)

	// Output:
	// val true
	//  false
}

func ExampleStore_Delete_multipleKeys() {
	store := memkv.New[string, string](0)

	key1, val1 := "key1", "val1"
	if err := store.Set(key1, val1); err != nil {
		return
	}

	key2, val2 := "key2", "val2"
	if err := store.Set(key2, val2); err != nil {
		return
	}

	v, ok := store.Get(key1)
	fmt.Println(v, ok)

	v, ok = store.Get(key2)
	fmt.Println(v, ok)

	store.Delete(key1, key2)

	v, ok = store.Get(key1)
	fmt.Println(v, ok)

	v, ok = store.Get(key2)
	fmt.Println(v, ok)

	// Output:
	// val1 true
	// val2 true
	//  false
	//  false
}

func ExampleStore_Flush() {
	store := memkv.New[string, string](0)

	key1, val1 := "key1", "val1"
	if err := store.Set(key1, val1); err != nil {
		return
	}

	key2, val2 := "key2", "val2"
	if err := store.Set(key2, val2); err != nil {
		return
	}

	l := store.Len()
	fmt.Println(l)

	store.Flush()

	l = store.Len()
	fmt.Println(l)

	// Output:
	// 2
	// 0
}

func ExampleStore_Len() {
	store := memkv.New[string, string](0)

	key, val := "key1", "val1"
	if err := store.Set(key, val); err != nil {
		return
	}

	l := store.Len()
	fmt.Println(l)

	if err := store.Set("key2", "val2"); err != nil {
		return
	}

	l = store.Len()
	fmt.Println(l)

	// Output:
	// 1
	// 2
}

func ExampleStore_Items() {
	store := memkv.New[string, string](0)

	key1, val1 := "key1", "val1"
	if err := store.Set(key1, val1); err != nil {
		return
	}

	key2, val2 := "key2", "val2"
	if err := store.Set(key2, val2); err != nil {
		return
	}

	items := store.Items()
	fmt.Println(items)

	// Output:
	// map[key1:val1 key2:val2]
}

func ExampleStore_Keys() {
	store := memkv.New[string, string](0)

	key1, val1 := "key1", "val1"
	if err := store.Set(key1, val1); err != nil {
		return
	}

	key2, val2 := "key2", "val2"
	if err := store.Set(key2, val2); err != nil {
		return
	}

	keys := store.Keys()
	fmt.Println(keys)

	// Output:
	// [key1 key2]
}

func ExampleStore_Values() {
	store := memkv.New[string, string](0)

	key1, val1 := "key1", "val1"
	if err := store.Set(key1, val1); err != nil {
		return
	}

	key2, val2 := "key2", "val2"
	if err := store.Set(key2, val2); err != nil {
		return
	}

	values := store.Values()
	fmt.Println(values)

	// Output:
	// [val1 val2]
}
