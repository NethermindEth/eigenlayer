package daemon

type Daemon interface {
	Install(options *InstallOptions) (*InstallResponse, error)
}
