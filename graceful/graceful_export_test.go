package graceful

// export for testing.
func (g *Group) ErrCh() chan error {
	if g.errCh == nil {
		g.errCh = make(chan error)
	}
	return g.errCh
}
