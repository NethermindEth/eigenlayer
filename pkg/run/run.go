package run

import (
	"github.com/NethermindEth/eigen-wiz/internal/package_handler"
)

// Runner is responsible for running a package that is already installed.
type Runner struct{}

// NewRunner creates a new Runner instance.
func NewRunner() *Runner {
	return &Runner{}
}

// Run starts the package at the given path.
func (r *Runner) Run(pkgPath string) error {
	pkgHandler := package_handler.NewPackageHandler(pkgPath)
	// TODO: check if the package is already running
	return pkgHandler.Run()
}
