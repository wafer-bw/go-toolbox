package always_test

import (
	"errors"
	"testing"

	"github.com/wafer-bw/go-toolbox/always"
)

func TestMust(t *testing.T) {
	t.Parallel()

	t.Run("returns result & does not panic when there is no error", func(t *testing.T) {
		t.Parallel()

		expectResult := "foo"
		input := func() (string, error) {
			return expectResult, nil
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("func always.Must(%T) should not panic\n\tPanic value: %v", input, r)
			}
		}()

		result := always.Must(input())
		if result != expectResult {
			t.Errorf("func always.Must(%T) = %v, want %v", input, result, expectResult)
		}
	})

	t.Run("panic when there is an error", func(t *testing.T) {
		t.Parallel()

		expectResult, expectPanic := "", "oh no"
		input := func() (string, error) {
			return expectResult, errors.New(expectPanic)
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("func always.Must(%T) should panic", input)
			} else if r.(error).Error() != expectPanic {
				t.Errorf("func always.Must(%T) should panic with %v, got %v", input, expectPanic, r)
			}
		}()

		result := always.Must(input())
		if result != expectResult {
			t.Errorf("func always.Must(%T) = %v, want %v", input, result, expectResult)
		}
	})
}

func TestMustDo(t *testing.T) {
	t.Parallel()

	t.Run("does not panic when there is no error", func(t *testing.T) {
		t.Parallel()

		input := func() error {
			return nil
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("func always.Must(%T) should not panic\n\tPanic value: %v", input, r)
			}
		}()

		always.MustDo(input())
	})

	t.Run("panic when there is an error", func(t *testing.T) {
		t.Parallel()

		expectPanic := "oh no"
		input := func() error {
			return errors.New(expectPanic)
		}

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("func always.Must(%T) should panic", input)
			} else if r.(error).Error() != expectPanic {
				t.Errorf("func always.Must(%T) should panic with %v, got %v", input, expectPanic, r)
			}
		}()

		always.MustDo(input())
	})
}

func TestAccept(t *testing.T) {
	t.Parallel()

	t.Run("returns the first argument, ignoring the first", func(t *testing.T) {
		t.Parallel()

		expect := "hello"
		if v := always.Accept(expect, "world"); v != expect {
			t.Errorf("expected %s, got %s", expect, v)
		}
	})
}
