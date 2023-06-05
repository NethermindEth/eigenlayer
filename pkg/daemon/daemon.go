package daemon

// Daemon is the main entrypoint for all the functionalities of the daemon.
type Daemon struct {
	installer Installer
}

// NewDaemon create a new daemon instance.
func NewDaemon(installer Installer) *Daemon {
	return &Daemon{
		installer: installer,
	}
}
