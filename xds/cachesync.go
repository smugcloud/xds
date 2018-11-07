package xds

type CacheEventHandler struct {
	Notifier
}

// Notifier supplies a callback to be called when changes occur

type Notifier interface {
	// OnChange is called to notify the callee that the
	// contents of the *dag.Builder have changed.
	OnChange(map[string]*Service)
}
