// Package funcopts provides a way to process functional options.
package funcopts

// Process executes the provided functional options on the passed pointer value.
func Process[V any, F ~func(*V) error](v *V, opts ...F) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}

		if err := opt(v); err != nil {
			return err
		}
	}

	return nil
}
