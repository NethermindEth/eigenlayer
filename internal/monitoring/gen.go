package monitoring

//go:generate mockgen -destination=./mocks/monitoring_service.go -package=mocks github.com/NethermindEth/egn/internal/monitoring ServiceAPI

//go:generate mockgen -destination=./mocks/compose.go -package=mocks github.com/NethermindEth/egn/internal/monitoring ComposeManager

//go:generate mockgen -destination=./mocks/docker.go -package=mocks github.com/NethermindEth/egn/internal/monitoring DockerManager
