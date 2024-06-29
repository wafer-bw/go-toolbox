package graceful

// export for testing.
func (g *Group) ErrCh() chan error {
	return g.errCh
}

// export for testing.
func (g *Group) CreateErrCh() {
	g.errCh = make(chan error)
}

func (g *Group) Waited() bool {
	return g.waited
}
