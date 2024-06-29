package graceful

// export for testing.
func (g *Group) ErrCh() chan error {
	return g.errCh
}

// export for testing.
func (g *Group) CreateErrCh(n int) {
	g.errCh = make(chan error, n)
}
