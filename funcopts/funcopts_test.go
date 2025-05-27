package funcopts_test

import (
	"errors"
	"testing"

	"github.com/wafer-bw/go-toolbox/funcopts"
)

func TestProcess(t *testing.T) {
	t.Parallel()

	t.Run("process all options", func(t *testing.T) {
		t.Parallel()

		v := &funcopts.T{}
		number := 1
		word := "hello"

		if err := funcopts.Process(v, funcopts.WithNumber(number), funcopts.WithWord(word)); err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if v.GetNumber() != number {
			t.Errorf("expected %d, got %d", number, v.GetNumber())
		}
		if v.GetWord() != word {
			t.Errorf("expected %s, got '%s'", word, v.GetWord())
		}
	})

	t.Run("no error on nil options", func(t *testing.T) {
		t.Parallel()

		v := &funcopts.T{}
		number := 1
		word := "hello"

		if err := funcopts.Process(v, funcopts.WithNumber(number), nil, funcopts.WithWord(word)); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("no error on empty options", func(t *testing.T) {
		t.Parallel()

		v := &funcopts.T{}

		if err := funcopts.Process[funcopts.T, funcopts.Option](v); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("returns error from option", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("option error")
		v := &funcopts.T{}
		o := func(v *funcopts.T) error { return expectedErr }

		if err := funcopts.Process(v, o); err == nil {
			t.Error("expected error, got nil")
		} else if err != expectedErr {
			t.Errorf("expected %s, got %s", expectedErr, err)
		}
	})
}
