package pointer_test

import (
	"testing"
	"time"

	"github.com/wafer-bw/go-toolbox/pointer"
)

func TestTo(t *testing.T) {
	t.Parallel()
	v := true
	pv := pointer.To(v)
	if pv == nil {
		t.Fatal("expected non-nil pointer")
	} else if v != *pv {
		t.Fatalf("expected %v, got %v", v, *pv)
	}
}

func TestFrom(t *testing.T) {
	t.Parallel()

	t.Run("not nil", func(t *testing.T) {
		t.Parallel()
		pv := new(bool)
		v := pointer.From(pv)
		if v != *pv {
			t.Fatalf("expected %v, got %v", *pv, v)
		}
	})

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		var e bool
		var pv *bool
		v := pointer.From(pv)
		if v != e {
			t.Fatalf("expected %v, got %v", e, v)
		}
	})
}

func TestToOrNil(t *testing.T) {
	t.Parallel()

	t.Run("zero value to nil", func(t *testing.T) {
		t.Parallel()
		var v string
		pv := pointer.ToOrNil(v)
		if pv != nil {
			t.Fatalf("expected nil pointer, got %v", pv)
		}
	})

	t.Run("non-zero value to pointer", func(t *testing.T) {
		t.Parallel()
		v := "non-zero"
		pv := pointer.ToOrNil(v)
		if pv == nil {
			t.Fatal("expected non-nil pointer")
		} else if v != *pv {
			t.Fatalf("expected %v, got %v", v, *pv)
		}
	})

	t.Run("non-zero value to pointer with IsZero", func(t *testing.T) {
		t.Parallel()
		v := time.Date(2014, 6, 25, 12, 24, 40, 0, time.UTC)
		pv := pointer.ToOrNil(v)
		if pv == nil {
			t.Fatal("expected non-nil pointer")
		} else if v != *pv {
			t.Fatalf("expected %v, got %v", v, *pv)
		}
	})

	t.Run("zero value to pointer with IsZero", func(t *testing.T) {
		t.Parallel()
		v := time.Time{}
		pv := pointer.ToOrNil(v)
		if pv != nil {
			t.Fatal("expected nil pointer")
		}
	})
}
