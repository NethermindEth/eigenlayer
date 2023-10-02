package daemon

//go:generate mockgen -destination=./mocks/monitoring_manager.go -package=mocks github.com/NethermindEth/eigenlayer/pkg/daemon MonitoringManager
//go:generate mockgen -destination=./mocks/compose.go -package=mocks github.com/NethermindEth/eigenlayer/pkg/daemon ComposeManager
//go:generate mockgen -destination=./mocks/docker.go -package=mocks github.com/NethermindEth/eigenlayer/pkg/daemon DockerManager
//go:generate mockgen -destination=./mocks/backup.go -package=mocks github.com/NethermindEth/eigenlayer/pkg/daemon BackupManager
