# Eigenlayer CLI

Eigenlayer is a setup wizard for EigenLayer Node Software. The tool installs, manages, and monitors EigenLayer nodes on your local machine. For more information on Eigenlayer, Eigenlayer Node Software, and what this tool does, check our [documentation](https://www.eigenlayer.xyz)

- [Eigenlayer CLI](#eigenlayer-cli)
  - [Dependencies](#dependencies)
  - [Install `eigenlayer` CLI using Go](#install-eigenlayer-cli-using-go)
  - [Install `eigenlayer` CLI from source](#install-eigenlayer-cli-from-source)
  - [Install `eigenlayer` CLI using a binary](#install-eigenlayer-cli-using-a-binary)
    - [Linux/amd64](#linuxamd64)
    - [Linux/arm64](#linuxarm64)
  - [Create and List Keys](#create-and-list-keys)
    - [Create keys](#create-keys)
    - [Import keys](#import-keys)
    - [List keys](#list-keys)
  - [Operator registration](#operator-registration)
    - [Sample config creation](#sample-config-creation)
  - [Install an AVS](#install-an-avs)
    - [From GitHub](#from-github)
    - [Non-interactive installation](#non-interactive-installation)
    - [From local directory](#from-local-directory)
  - [Updating AVS Node](#updating-avs-node)
    - [Updating with explicit version](#updating-with-explicit-version)
    - [Updating with commit hash](#updating-with-commit-hash)
    - [Updating options](#updating-options)
  - [Backup](#backup)
  - [List backups](#list-backups)
  - [Restore](#restore)
  - [Uninstalling AVS Node Software](#uninstalling-avs-node-software)
  - [List installed instances](#list-installed-instances)
  - [Run an AVS instance](#run-an-avs-instance)
  - [Stop an AVS instance](#stop-an-avs-instance)
  - [Logs](#logs)
  - [Init Monitoring Stack](#init-monitoring-stack)
  - [Clean Up Monitoring Stack](#clean-up-monitoring-stack)
  - [Running a Plugin](#running-a-plugin)
    - [Passing arguments to the plugin](#passing-arguments-to-the-plugin)

## Dependencies

This tool depends on [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) in order to manage the installation and running of EigenLayer nodes. Please make sure that you have both Docker and Docker Compose installed and configured properly before using this tool.

## Install `eigenlayer` CLI using Go

First, install the Go programming language following the [official instructions](https://go.dev/doc/install). You need at least the `1.21` version.

> Eigenlayer is only supported on **Linux**. Make sure you install Go for Linux in a Linux environment (e.g. WSL2, Docker, etc.)

This command will install the `eigenlayer` executable along with the library and its dependencies in your system:

> As the repository is private, you need to set the `GOPRIVATE` variable properly by running the following command: `export GOPRIVATE=github.com/NethermindEth/eigenlayer,$GOPRIVATE`. Git will automatically resolve the private access if your Git user has all the required permissions over the repository.

```bash
go install github.com/NethermindEth/eigenlayer/cmd/eigenlayer@latest
```

The executable will be in your `$GOBIN` (`$GOPATH/bin`).

To check if the `GOBIN` is not in your PATH, you can execute `echo $GOBIN` from the Terminal. If it doesn't print anything, then it is not in your PATH. To add `GOBIN` to your PATH, add the following lines to your `$HOME/.profile`:

```bash
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
```

> Changes made to a profile file may not apply until the next time you log into your computer. To apply the changes immediately, run the shell commands directly or execute them from the profile using a command such as `source $HOME/.profile`.

## Install `eigenlayer` CLI from source

With this method, you generate the binary manually (need Go installed), downloading and compiling the source code:

```bash
git clone https://github.com/NethermindEth/eigenlayer.git
cd eigenlayer
mkdir -p build
go build -o build/eigenlayer cmd/eigenlayer/main.go
```

or if you have `make` installed:

```bash
git clone https://github.com/NethermindEth/eigenlayer.git
cd eigenlayer
make build
```

The executable will be in the `build` folder.

---
In case you want the binary in your PATH (or if you used the [Using Go](#install-eigenlayer-cli-using-go) method and you don't have `$GOBIN` in your PATH), please copy the binary to `/usr/local/bin`:

```bash
# Using Go
sudo cp $GOPATH/bin/eigenlayer /usr/local/bin/

# Build from source
sudo cp eigenlayer/build/eigenlayer /usr/local/bin/
```

## Install `eigenlayer` CLI using a binary

The `eigenlayer` CLI tool versions are managed with GitHub releases. To install it, you can download the binary directly from the release assets manually, or by using the following command replacing the `<VERSION>` and `<ARCH>` with the proper values:

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/<VERSION>/eigenlayer-linux-<ARCH> --output eigenlayer
```

### Linux/amd64

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/v0.2.1/eigenlayer-linux-amd64 --output eigenlayer
```

### Linux/arm64

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/v0.2.1/eigenlayer-linux-arm64 --output eigenlayer
```

## Create and List Keys

### Create keys

You can create encrypted ecdsa and bls keys using the cli which will be needed for operator registration and other onchain calls

```bash
eigenlayer operator keys create --key-type ecdsa [keyname]
eigenlayer operator keys create --key-type bls [keyname]
```
- `keyname` - This will be the name of the created key file. It will be saved as `<keyname>.ecdsa.key.json` or `<keyname>.bls.key.json`

This will prompt a password which you can use to encrypt the keys. Keys will be stored in local disk and will be shown once keys are created.
It will also show the private key only once, so that you can back it up in case you lose the password or keyfile.

Example:

Input command
```bash
eigenlayer operator keys create --key-type ecdsa test
```
Output
```bash
? Enter password to encrypt the ecdsa private key: *******
ECDSA Private Key (Hex):  6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6
Please backup the above private key hex in safe place.

Key location: ./operator_keys/test.ecdsa.key.json
a30264c19cd7292d5153da9c9df58f81aced417e8587dd339021c45ee61f20d55f4c3d374d6f472d3a2c4382e2a9770db395d60756d3b3ea97e8c1f9013eb1bb
0x9F664973BF656d6077E66973c474cB58eD5E97E1
```

### Import keys

You can import existing ecdsa and bls keys using the cli which will be needed for operator registration and other onchain calls

```bash
eigenlayer operator keys import --key-type ecdsa [keyname] [privatekey]
eigenlayer operator keys import --key-type bls [keyname] [privatekey]
```
- `keyname` - This will be the name of the imported key file. It will be saved as `<keyname>.ecdsa.key.json` or `<keyname>.bls.key.json`
- `privatekey` - This will be the private key of the key to be imported.
  - For ecdsa key, it should be in hex format
  - For bls key, it should be a large number

Example:

Input command
```bash
eigenlayer operator keys import --key-type ecdsa test 6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6
```
Output
```bash
? Enter password to encrypt the ecdsa private key: *******
ECDSA Private Key (Hex):  6842fb8f5fa574d0482818b8a825a15c4d68f542693197f2c2497e3562f335f6
Please backup the above private key hex in safe place.

Key location: ./operator_keys/test.ecdsa.key.json
a30264c19cd7292d5153da9c9df58f81aced417e8587dd339021c45ee61f20d55f4c3d374d6f472d3a2c4382e2a9770db395d60756d3b3ea97e8c1f9013eb1bb
0x9F664973BF656d6077E66973c474cB58eD5E97E1
```

This will prompt a password which you can use to encrypt the keys. Keys will be stored in local disk and will be shown once keys are created.
It will also show the private key only once, so that you can back it up in case you lose the password or keyfile.

### List keys

You can also list you created key using

```bash
eigenlayer operator keys list
```

It will show all the keys created with this command with the public key

## Operator registration

You can register your operator using the below command

```bash
eigenlayer operator register operator-config.yaml
```

A sample yaml [config file](cli/operator/config/operator-config-example.yaml) and [metadata](cli/operator/config/metadata-example.json) is provided for reference. You can also create empty config files by using commands referred in [this section](#sample-config-creation). Fill in the required details to register the operator.
Make sure that if you use `local_keystore` as signer, you give the path to the keys created in above section.

You can check the registration status of your operator using

```bash
eigenlayer operator status operator-config.yaml
```

You can also update the operator metadata using

```bash
eigenlayer operator update operator-config.yaml
```

### Sample config creation

If you need to create a new config file for registration and metadata you can use

```bash
eigenlayer operator config create
```

It will create two file: `operator.yaml` and `metadata.json`
After filling the details in `metadata.json`, please upload this into a publicly accessible location and fill that url in `operator.yaml`. A valid metadata url is required for successful registration.

## Install an AVS

AVS Node software can be installed from a Git repository, such as GitHub, or from a local directory, as long as the package adheres to the packaging specification.

> Each AVS profiles defines a set of services using Docker Compose. For security reasons, building Docker images is not supported, and profiles cannot use the `build` option in their `docker-compose.yml` files. Only the `image` option is supported. This applies to the plugin as well. In the `manifest.yml` file, the `plugin` option only supports the `image` field, which should contain a Docker image name as its value, rather than a reference to a Dockerfile.

Further details on how to proceed in each case are provided below.

### From GitHub

We have created a mock-avs-pkg repository to understand the structure of an AVS Node Software package and to test all the features of the `eigenlayer` CLI tool. The following command shows how to install `mock-avs-pkg` using the `eigenlayer` tool:

```bash
eigenlayer install https://github.com/NethermindEth/mock-avs-pkg
```

Executing this command triggers an interactive installation process. During this process, the user can manually select the desired profile and all necessary options. Below is the final output after options have been selected.

```bash
INFO[0000] Version not specified, using latest.
INFO[0000] Using version v5.4.0
? Select a profile option-returner
? main-container-name option-returner
? main-port 8080
? network-name eigenlayer
? test-option-int 666
? test-option-float 666.666
? test-option-bool true
? test-option-path-dir /tmp
? test-option-path-file /tmp/test.txt
? test-option-uri https://www.google.com
? test-option-enum option1
INFO[0004] Installed successfully with instance id: mock-avs-default
INFO[0004] The installed node software has a plugin.
? Run the new instance now? No
```

### Non-interactive installation

To skip the interactive installation, the user can use the available flags of the `install` command. To see all available `install` options, run the `eigenlayer install --help` command. This is an example of installing the same `mock-avs` without interactive installation:

```bash
$ eigenlayer install \
 --profile option-returner \
 --no-prompt \
 https://github.com/NethermindEth/mock-avs-pkg
```

Output:

```bash
INFO[0000] Version not specified, using latest.
INFO[0000] Using version v5.4.0
INFO[0002] Installed successfully with instance id: mock-avs-default
INFO[0002] The installed node software has a plugin.
```

Notice the usage of:

- `--profile` to select the `option-returner` profile without prompt.
- `--no-prompt` to skip options prompts.

In this case, the `option-returner` profile uses all the default values. To set option values, use the `--option.<option-name>` dynamic flags. For instance:

```bash
$ eigenlayer install \
 --profile option-returner \
 --no-prompt \
 --option.main-port 8081 \
 https://github.com/NethermindEth/mock-avs-pkg
```

Output:

```bash
INFO[0000] Version not specified, using latest.
INFO[0000] Using version v5.4.0
INFO[0002] Installed successfully with instance id: mock-avs-default
INFO[0002] The installed node software has a plugin.
```

In this case, the `main-port` has a value of 8081 instead of the default value of 8080.

### From local directory

> THIS INSTALLATION METHOD IS INSECURE
>

Installing from a local directory can be helpful for AVS developers who want to test Node Software packaging before releasing it to a public Git repository. To install an AVS Node Software from a local directory, use the `eigenlayer local-install` command. To illustrate local installation, let's clone the `mock-avs-pkg` to a local directory, and use it as a local package.

First, clone the `mock-avs-pkg` package:

> If you already have a local package, you can skip this step

```bash
git clone --branch v5.4.0 https://github.com/NethermindEth/mock-avs-pkg
```

Now we can install the package from the `mock-avs-pkg` directory with the following command:

```bash
eigenlayer local-install ./mock-avs-pkg --profile option-returner
```

Output:

```bash
WARN[0000] This command is insecure and should only be used for development purposes
INFO[0001] Installed successfully with instance id: mock-avs-default
```

## Updating AVS Node

To update an installed AVS Node Software, use the `eigenlayer update` command with the AVS instance ID as an argument, as follows:

```bash
eigenlayer update mock-avs-default
```

Output:

```log
INFO[0000] Pulling package...                           
INFO[0000] Package pulled successfully                  
INFO[0000] Package version changed: v5.4.0 -> v5.5.0    
INFO[0000] Package commit changed from b64c50c15e53ae7afebbdbe210b834d1ee471043 -> a3406616b848164358fdd24465b8eecda5f5ae34 
INFO[0000] Uninstalling current package...              
INFO[0000] Package uninstalled successfully             
INFO[0000] Installing new package...                    
INFO[0000] Package installed successfully with instance ID: mock-avs-default 
INFO[0000] The installed node software has a plugin.    
INFO[0000] Instance mock-avs-default running successfully
```

> This case updates the instance to the latest version, if possible. If the latest version is already installed, then the command does not do anything.

### Updating with explicit version

To update to an specific version, pass the version as an argument. The new version mus be greater than the current version following the [semver](https://semver.org/) specification. For instance:

```bash
eigenlayer update mock-avs-default v5.5.0
```

### Updating with commit hash

To update to a specific commit, pass the commit hash as an argument. The new commit should be a descendant of the installed commit, we guarantee that checking the git history from the pulled package. For instance:

```bash
eigenlayer update mock-avs-default a3406616b848164358fdd24465b8eecda5f5ae34
```

### Updating options

The `--no-prompt` flag is available to skip the options prompt, also the dynamic flags `--option.<option-name>` are available to set the option values, like in the `install` command.

## Backup

To backup an installed AVS Node Software, use the `eigenlayer backup` command with the AVS instance ID as an argument, as follows:

```bash
eigenlayer backup mock-avs-default
```

Output:

```log
INFO[0000] Backing up instance mock-avs-default         
INFO[0000] Backing up instance data...                  
INFO[0000] Backup created with id: mock-avs-default-1696337650
```

## List backups

To list all the backups, use the `eigenlayer backup ls` command, as follows:

```bash
eigenlayer backup ls
```

Output:

```bash
ID          AVS Instance ID     VERSION    COMMIT                                      TIMESTAMP              SIZE     URL                                              
6ee67470    mock-avs-default    v5.5.1     d5af645fffb93e8263b099082a4f512e1917d0af    2023-10-04 13:41:06    10KiB    https://github.com/NethermindEth/mock-avs-pkg
```

## Restore

To restore a backup, use the `eigenlayer restore` command with the backup ID as an argument, as follows:

```bash
eigenlayer restore <backup-id>
```

If the AVS instance id of the backup exists, then the command will uninstall it before restoring the backup. If the AVS instance does not exist, then the command will create it. To run the restored instance after the restore process, use the `--run` flag as follows:

```bash
eigenlayer restore --run <backup-id>
```

## Uninstalling AVS Node Software

When uninstalling AVS Node Software, it is stopped, disconnected from the Monitoring Stack, and removed from the data directory. To uninstall AVS Node Software, use the `eigenlayer uninstall` command with the AVS instance ID as an argument, as follows:

```bash
eigenlayer uninstall mock-avs-default
```

> To see the ID of all installed instances, use the `eigenlayer ls` command.

## List installed instances

To list the installed instances, their status, and health, run the command `eigenlayer ls`. For example:

```bash
eigenlayer ls
```

Output:

```bash
AVS Instance ID     RUNNING    HEALTH     COMMENT
mock-avs-default    true       healthy
```

## Run an AVS instance

To run a stopped instance, use the `eigenlayer run` command with the instance ID as an argument in the following format:

```bash
eigenlayer run mock-avs-default
```

If the instance is already running, then the command does not do anything.

## Stop an AVS instance

To stop a running instance, use the `eigenlayer stop` command with the instance ID as an argument in the following format:

```bash
eigenlayer stop mock-avs-default
```

## Logs

AVS instance logs could be retrieved using the `eigenlayer logs` command. Logs are the merge of docker containers logs that compounds the AVS instance, for instance:

```bash
eigenlayer logs mock-avs-default
```

```log
option-returner: INFO:     Started server process [1]
option-returner: INFO:     Waiting for application startup.
option-returner: INFO:     Application startup complete.
option-returner: INFO:     Uvicorn running on <http://0.0.0.0:8080> (Press CTRL+C to quit)
option-returner: INFO:     172.20.0.3:59224 - "GET /metrics HTTP/1.1" 307 Temporary Redirect
option-returner: INFO:     172.20.0.3:59224 - "GET / HTTP/1.1" 200 OK
option-returner: INFO:     172.20.0.3:40780 - "GET /metrics HTTP/1.1" 307 Temporary Redirect
option-returner: INFO:     172.20.0.3:40780 - "GET / HTTP/1.1" 200 OK
```

## Init Monitoring Stack

To initialize the monitoring stack, use the command `eigenlayer init-monitoring`. After that, you can check if Grafana is running on [http://localhost:3000](http://localhost:3000/) and if the Prometheus server is running on [http://localhost:9090](http://localhost:9090/).

## Clean Up Monitoring Stack

To stop and clean up the Monitoring Stack, run the command `eigenlayer clean-monitoring`.

## Running a Plugin

An AVS package may include a plugin that is detected during the AVS installation. To run a plugin, you need to know the ID of the instance that has the desired plugin. You can obtain this information by running the `eigenlayer ls` command beforehand. For example:

```bash
eigenlayer ls
```

Output:

```bash
AVS Instance ID     RUNNING    HEALTH     COMMENT
mock-avs-default    true       healthy
```

To run the plugin of the `mock-avs-default` instance, use the following command:

```bash
eigenlayer plugin mock-avs-default
```

Output:

```log
INFO[0001] Running plugin with image eigen-plugin-mock-avs-default on network eigenlayer
INFO[0002]
AVS is up
```

This command will run the plugin container inside the AVS Docker network and remove it after execution.

The `--host` flag can be used to run the plugin container in the `host` network if the AVS Node Software is not running. To mount directories, files, and Docker volumes to the plugin container, use the `--volume` flag. All `plugin` command flags should be declared before the instance ID. To learn more about these options, run the `eigenlayer plugin --help` command.

### Passing arguments to the plugin

To pass arguments to the plugin container ENTRYPOINT, append them after the AVS instance ID, as follows:

```bash
eigenlayer plugin mock-avs-default --port 8080
```

Output:

```bash
INFO[0004] Running plugin with image eigen-plugin-mock-avs-default on network eigenlayer
INFO[0004]
AVS is up
```

In this case, the plugin container receives the `--port 8080` arguments. Note that this is not a flag of the `eigenlayer plugin` command.