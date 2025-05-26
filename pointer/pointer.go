package pointer

type IsZeroer interface {
	IsZero() bool
}

// To returns a pointer to v.
func To[T any](v T) *T {
	return &v
}

// ToOrNil returns a pointer to v if it is not zero or nil otherwise.
//
// If v implements the [IsZeroer] interface it will use that to determine if it
// is zero.
func ToOrNil[T comparable](v T) *T {
	if z, ok := any(v).(IsZeroer); ok {
		if z.IsZero() {
			return nil
		}
		return &v
	}

	var z T
	if v == z {
		return nil
	}

	return &v
}

// From returns the value pointed at by p or the zero value of p's type if it is
// nil.
func From[T any](p *T) T {
	if p == nil {
		var z T
		return z
	}

	return *p
}
