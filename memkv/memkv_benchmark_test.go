package memkv_test

import (
	"fmt"
	"testing"

	"github.com/wafer-bw/go-toolbox/memkv"
)

var sizes = []int{100, 1000, 10000, 100000}

func BenchmarkStore_Set(b *testing.B) {
	for _, size := range sizes {
		store := memkv.New[int, int](size)
		for i := 0; i < size; i++ {
			if err := store.Set(i, i); err != nil {
				b.Fatal(err)
			}
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if err := store.Set(i%size, i); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkStore_Get(b *testing.B) {
	for _, size := range sizes {
		store := memkv.New[int, int](size)
		for i := 0; i < size; i++ {
			if err := store.Set(i, i); err != nil {
				b.Fatal(err)
			}
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				v, ok := store.Get(i % size)
				_, _ = v, ok
			}
		})
	}
}

func BenchmarkStore_Delete(b *testing.B) {
	for _, size := range sizes {
		store := memkv.New[int, int](size)
		for i := 0; i < size; i++ {
			if err := store.Set(i, i); err != nil {
				b.Fatal(err)
			}
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				store.Delete(i % size)
			}
		})
	}
}

func BenchmarkStore_Flush(b *testing.B) {
	for _, size := range sizes {
		store := memkv.New[int, int](size)
		for i := 0; i < size; i++ {
			if err := store.Set(i, i); err != nil {
				b.Fatal(err)
			}
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				store.Flush()
			}
		})
	}
}

func BenchmarkStore_Len(b *testing.B) {
	for _, size := range sizes {
		store := memkv.New[int, int](size)
		for i := 0; i < size; i++ {
			if err := store.Set(i, i); err != nil {
				b.Fatal(err)
			}
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				l := store.Len()
				_ = l
			}
		})
	}
}

func BenchmarkStore_Items(b *testing.B) {
	for _, size := range sizes {
		store := memkv.New[int, int](size)
		for i := 0; i < size; i++ {
			if err := store.Set(i, i); err != nil {
				b.Fatal(err)
			}
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				items := store.Items()
				_ = items
			}
		})
	}
}

func BenchmarkStore_Keys(b *testing.B) {
	for _, size := range sizes {
		store := memkv.New[int, int](size)
		for i := 0; i < size; i++ {
			if err := store.Set(i, i); err != nil {
				b.Fatal(err)
			}
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				keys := store.Keys()
				_ = keys
			}
		})
	}
}

func BenchmarkStore_Values(b *testing.B) {
	for _, size := range sizes {
		store := memkv.New[int, int](size)
		for i := 0; i < size; i++ {
			if err := store.Set(i, i); err != nil {
				b.Fatal(err)
			}
		}

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				values := store.Values()
				_ = values
			}
		})
	}
}
