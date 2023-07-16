package cli

//go:generate mockgen -destination=./mocks/daemon.go -package=mocks github.com/NethermindEth/eigenlayer/pkg/daemon Daemon
//go:generate mockgen -destination=./mocks/option.go -package=mocks github.com/NethermindEth/eigenlayer/pkg/daemon Option
