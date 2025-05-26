package pointer_test

import (
	"testing"
	"time"

	"github.com/wafer-bw/go-toolbox/pointer"
)

func BenchmarkTo(b *testing.B) {
	v := true
	for b.Loop() {
		pv := pointer.To(v)
		_ = pv
	}
}

func BenchmarkToOrNil(b *testing.B) {
	b.Run("zero value to nil", func(b *testing.B) {
		var v string
		for b.Loop() {
			pv := pointer.ToOrNil(v)
			_ = pv
		}
	})

	b.Run("non-zero value to pointer", func(b *testing.B) {
		v := "non-zero"
		for b.Loop() {
			pv := pointer.ToOrNil(v)
			_ = pv
		}
	})

	b.Run("non-zero value to pointer with IsZero", func(b *testing.B) {
		v := time.Date(2014, 6, 25, 12, 24, 40, 0, time.UTC)
		for b.Loop() {
			pv := pointer.ToOrNil(v)
			_ = pv
		}
	})

	b.Run("zero value to pointer with IsZero", func(b *testing.B) {
		v := time.Time{}
		for b.Loop() {
			pv := pointer.ToOrNil(v)
			_ = pv
		}
	})
}

func BenchmarkFrom(b *testing.B) {
	b.Run("not nil", func(b *testing.B) {
		pv := new(bool)
		for b.Loop() {
			v := pointer.From(pv)
			_ = v
		}
	})

	b.Run("nil", func(b *testing.B) {
		var pv *bool
		for b.Loop() {
			v := pointer.From(pv)
			_ = v
		}
	})
}
