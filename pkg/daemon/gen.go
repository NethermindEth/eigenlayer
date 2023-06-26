package daemon

//go:generate mockgen -destination=./mocks/monitoring_manager.go -package=mocks github.com/NethermindEth/egn/pkg/daemon MonitoringManager
//go:generate mockgen -destination=./mocks/compose.go -package=mocks github.com/NethermindEth/egn/pkg/daemon ComposeManager
