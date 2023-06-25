package daemon

//go:generate mockgen -destination=./mocks/daemon.go -package=mocks github.com/NethermindEth/egn/pkg/daemon Daemon
//go:generate mockgen -destination=./mocks/option.go -package=mocks github.com/NethermindEth/egn/pkg/daemon Option
//go:generate mockgen -destination=./mocks/monitoring_manager.go -package=mocks github.com/NethermindEth/egn/pkg/daemon MonitoringManager
