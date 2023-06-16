package daemon

type Daemon interface {
	Install(options InstallOptions) error
}
