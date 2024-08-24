package memkv_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wafer-bw/go-toolbox/memkv"
	"github.com/wafer-bw/go-toolbox/memkv/internal/underlying"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("creates a new store with no capacity", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)
		require.Zero(t, store.Capacity())
	})

	t.Run("creates a new store with a capacity", func(t *testing.T) {
		t.Parallel()

		capacity := 10
		store := memkv.New[string, string](capacity)
		require.NotNil(t, store)
		require.Equal(t, capacity, store.Capacity())
	})

	t.Run("creates a new store with no capacity when provided a negative capacity", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](-1)
		require.NotNil(t, store)
		require.Zero(t, store.Capacity())
	})
}

func TestStore_Set(t *testing.T) {
	t.Parallel()

	t.Run("sets a value in the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)

		key, val := "key", "val"

		err := store.Set(key, val)
		require.NoError(t, err)

		data, unlock := store.Data()
		defer unlock()
		require.Len(t, data.Items, 1)

		item, ok := data.Items[key]
		require.True(t, ok)
		require.Equal(t, val, item.Value)
	})

	t.Run("returns an error when the store is at capacity", func(t *testing.T) {
		t.Parallel()

		capacity := 1
		store := memkv.New[string, string](capacity)
		require.NotNil(t, store)

		key1, val1 := "key1", "val1"
		key2, val2 := "key2", "val2"

		err := store.Set(key1, val1)
		require.NoError(t, err)

		err = store.Set(key2, val2)
		require.Error(t, err)
		require.IsType(t, &memkv.AtCapacityError{}, err)
		require.Equal(t, err.Error(), "store is at capacity")
	})
}

func TestStore_Get(t *testing.T) {
	t.Parallel()

	t.Run("gets a value from the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)

		key, val := "key", "val"

		data, unlock := store.Data()
		data.Items[key] = underlying.Item[string, string]{Value: val}
		unlock()

		v, ok := store.Get(key)
		require.True(t, ok)
		require.Equal(t, val, v)
	})

	t.Run("returns false when the key is not in the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)

		_, ok := store.Get("key")
		require.False(t, ok)
	})
}

func TestStore_Delete(t *testing.T) {
	t.Parallel()

	t.Run("deletes a value from the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)

		key, val := "key", "val"

		data, unlock := store.Data()
		data.Items[key] = underlying.Item[string, string]{Value: val}
		unlock()

		store.Delete(key)

		data, unlock = store.Data()
		defer unlock()
		require.Len(t, data.Items, 0)
	})

	t.Run("deletes multiple values from the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)

		key1, val1 := "key1", "val1"
		key2, val2 := "key2", "val2"

		data, unlock := store.Data()
		data.Items[key1] = underlying.Item[string, string]{Value: val1}
		data.Items[key2] = underlying.Item[string, string]{Value: val2}
		unlock()

		store.Delete(key1, key2)

		data, unlock = store.Data()
		defer unlock()
		require.Len(t, data.Items, 0)
	})

	t.Run("nothing happens when the keys are not in the store", func(t *testing.T) {
		t.Parallel()

		require.NotPanics(t, func() {
			store := memkv.New[string, string](0)
			require.NotNil(t, store)

			store.Delete("key1", "key2")
		})
	})
}

func TestStore_Flush(t *testing.T) {
	t.Parallel()

	t.Run("flushes the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)

		key1, val1 := "key1", "val1"
		key2, val2 := "key2", "val2"

		data, unlock := store.Data()
		data.Items[key1] = underlying.Item[string, string]{Value: val1}
		data.Items[key2] = underlying.Item[string, string]{Value: val2}
		unlock()

		store.Flush()

		data, unlock = store.Data()
		defer unlock()
		require.Len(t, data.Items, 0)
	})

	t.Run("nothing happens when the store is empty", func(t *testing.T) {
		t.Parallel()

		require.NotPanics(t, func() {
			store := memkv.New[string, string](0)
			require.NotNil(t, store)

			store.Flush()
		})
	})
}

func TestStore_Len(t *testing.T) {
	t.Parallel()

	t.Run("returns the number of items in the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)
		require.Zero(t, store.Len())

		key1, val1 := "key1", "val1"
		key2, val2 := "key2", "val2"

		data, unlock := store.Data()
		data.Items[key1] = underlying.Item[string, string]{Value: val1}
		data.Items[key2] = underlying.Item[string, string]{Value: val2}
		unlock()

		require.Equal(t, 2, store.Len())
	})
}

func TestStore_Items(t *testing.T) {
	t.Parallel()

	t.Run("returns the items in the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)
		require.Len(t, store.Items(), 0)

		key1, val1 := "key1", "val1"
		key2, val2 := "key2", "val2"

		data, unlock := store.Data()
		data.Items[key1] = underlying.Item[string, string]{Value: val1}
		data.Items[key2] = underlying.Item[string, string]{Value: val2}
		unlock()

		items := store.Items()
		require.Len(t, items, 2)
		v1, ok := items[key1]
		require.True(t, ok)
		require.Equal(t, val1, v1)

		v2, ok := items[key2]
		require.True(t, ok)
		require.Equal(t, val2, v2)
	})
}

func TestStore_Keys(t *testing.T) {
	t.Parallel()

	t.Run("returns the keys in the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)
		require.Len(t, store.Keys(), 0)

		key1, val1 := "key1", "val1"
		key2, val2 := "key2", "val2"

		data, unlock := store.Data()
		data.Items[key1] = underlying.Item[string, string]{Value: val1}
		data.Items[key2] = underlying.Item[string, string]{Value: val2}
		unlock()

		keys := store.Keys()
		require.Len(t, keys, 2)
		require.Contains(t, keys, key1)
		require.Contains(t, keys, key2)
	})
}

func TestStore_Values(t *testing.T) {
	t.Parallel()

	t.Run("returns the values in the store", func(t *testing.T) {
		t.Parallel()

		store := memkv.New[string, string](0)
		require.NotNil(t, store)
		require.Len(t, store.Values(), 0)

		key1, val1 := "key1", "val1"
		key2, val2 := "key2", "val2"

		data, unlock := store.Data()
		data.Items[key1] = underlying.Item[string, string]{Value: val1}
		data.Items[key2] = underlying.Item[string, string]{Value: val2}
		unlock()

		values := store.Values()
		require.Len(t, values, 2)
		require.Contains(t, values, val1)
		require.Contains(t, values, val2)
	})
}
