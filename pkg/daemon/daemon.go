package daemon

type Daemon interface {
	Pull(options *PullOptions) (*PullResponse, error)
	Install(options *InstallOptions) (*InstallResponse, error)
}
