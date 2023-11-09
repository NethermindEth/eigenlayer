# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.4.3] 2023-11-08
- support for ubuntu 20.04 binaries ([#140](https://github.com/NethermindEth/eigenlayer/pull/140))

## [v0.4.2] 2023-11-05
- dependency updated ([#137](https://github.com/NethermindEth/eigenlayer/pull/137))

## [v0.4.1] - 2023-10-31
- Update eigensdk - support only png images for metadata ([#134](https://github.com/NethermindEth/eigenlayer/pull/134)
- prompt operator for creating config files ([#125](https://github.com/NethermindEth/eigenlayer/pull/125))
- Print etherscan link for on-chain tx ([#124](https://github.com/NethermindEth/eigenlayer/pull/124))

## [v0.4.0] - 2023-10-24
- Commented the install and metrics related operations in the eigenlayer command ([#127](https://github.com/NethermindEth/eigenlayer/pull/127))

## [v0.3.1] - 2023-10-19
### Fixed
- Updated eigensdk to [v0.0.6](https://github.com/Layr-Labs/eigensdk-go/releases/tag/v0.0.6) ([#122](https://github.com/NethermindEth/eigenlayer/pull/122)) 
## [v0.3.0] - 2023-10-19
### Added

- Add `local-update` command. ([#98](https://github.com/NethermindEth/eigenlayer/pull/98))
- Updated eigensdk to [v0.0.5](https://github.com/Layr-Labs/eigensdk-go/releases/tag/v0.0.5) ([#120](https://github.com/NethermindEth/eigenlayer/pull/120))

## [v0.2.1] - 2023-10-11 

### Fixed

- Pull plugin image if does not exist locally ([#101](https://github.com/NethermindEth/eigenlayer/pull/101))
- Updated keystore folder relative to user's HOME directory ([#118](https://github.com/NethermindEth/eigenlayer/pull/118))

## [v0.2.0] - 2023-10-10

### Added

- Sort backup `ls` command results by date. ([#95](https://github.com/NethermindEth/eigenlayer/pull/95))
- Enforce and validate password on key creation. ([#89](https://github.com/NethermindEth/eigenlayer/pull/89))
- Add `restore` command. ([#90](https://github.com/NethermindEth/eigenlayer/pull/90))
  - Upgrade `update` command to support backing up the old instance and restoring from a backup if the update fails.
- Add `keys import` command  ([#97](https://github.com/NethermindEth/eigenlayer/pull/97))

### Fixed

- Update and review Release pipeline ppa packaging. ([#94](https://github.com/NethermindEth/eigenlayer/pull/94))

## [v0.1.0] - 2023-10-05

### Added

- Initial CLI with the following commands:
  - `install` command to install an AVS node from a given Git repository.
  - `local-install` command to install an AVS node from a local directory. This is useful for development purposes, and it is not intended to be used in production.
  - `run` command to run an AVS node instance that has been installed.
  - `stop` command to stop an AVS node instance.
  - `ls` command to list the installed AVS nodes and their status.
  - `logs` command to show the logs of an AVS node instance.
  - `uninstall` command to uninstall an AVS node instance.
  - `plugin` command to run an AVS node plugin.
  - `clean-monitoring` command to clean the Monitoring Stack.
  - `init-monitoring` command to initialize the Monitoring Stack.
  - `update` command to update an AVS node instance.
  - `backup` command to backup an AVS node instance.
  - `operator` command to access utilities for the AVS Operator such as keys management, status, register, etc, using eigenSDK.
- Full operational monitoring stack with Prometheus, Node Exporter, and Grafana.
- Support for the [Eigenlayer AVS Node Specification v0.1.0](https://eigen.nethermind.io/).

<!-- ### Fixed -->

<!-- ### Changed -->

<!-- ### Removed -->
