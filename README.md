# Eigenlayer CLI

Eigenlayer is a setup wizard for EigenLayer Node Software. The tool installs, manages, and monitors EigenLayer nodes on your local machine.


- [Eigenlayer CLI](#eigenlayer-cli)
  - [Install `eigenlayer` CLI](#install-eigenlayer-cli)
    - [Linux/amd64](#linuxamd64)
    - [Linux/arm64](#linuxarm64)
    - [Dependencies](#dependencies)
  - [Install an AVS](#install-an-avs)
    - [From GitHub](#from-github)
    - [Non-interactive installation](#non-interactive-installation)
    - [From local directory](#from-local-directory)
  - [Uninstalling AVS Node Software](#uninstalling-avs-node-software)
  - [List installed instances](#list-installed-instances)
  - [Run an AVS instance](#run-an-avs-instance)
  - [Stop an AVS instance](#stop-an-avs-instance)
  - [Logs](#logs)
  - [Init Monitoring Stack](#init-monitoring-stack)
  - [Clean Up Monitoring Stack](#clean-up-monitoring-stack)
  - [Running a Plugin](#running-a-plugin)
    - [Passing arguments to the plugin](#passing-arguments-to-the-plugin)

## Install `eigenlayer` CLI

The `eigenlayer` CLI tool versions are managed with GitHub releases. To install it, you can download the binary directly from the release assets manually, or by using the following command replacing the `<VERSION>` and `<ARCH>` with the proper values:

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/<VERSION>/eigenlayer-linux-<ARCH> --output eigenlayer
```

### Linux/amd64

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/v0.1.0/eigenlayer-linux-amd64 --output eigenlayer
```

### Linux/arm64

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/v0.1.0/eigenlayer-linux-arm64 --output eigenlayer
```

### Dependencies

**Note:** This tool depends on [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) in order to manage the installation and running of EigenLayer nodes. Please make sure that you have both Docker and Docker Compose installed and configured properly before using this tool.

## Install an AVS

AVS Node software can be installed from a Git repository, such as GitHub, or from a local directory, as long as the package adheres to the packaging specification. Further details on how to proceed in each case are provided below.

### From GitHub

We have created a mock-avs repository to understand the structure of an AVS Node Software package and to test all the features of the `eigenlayer` CLI tool. The following command shows how to install `mock-avs` using the `eigenlayer` tool:

```bash
eigenlayer install https://github.com/NethermindEth/mock-avs
```

Executing this command triggers an interactive installation process. During this process, the user can manually select the desired profile and all necessary options. Below is the final output after options have been selected.

```bash
INFO[0000] Version not specified, using latest.
INFO[0000] Using version v3.1.0
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
 https://github.com/NethermindEth/mock-avs
```

Output:

```bash
INFO[0000] Version not specified, using latest.
INFO[0000] Using version v3.1.0
INFO[0002] Installed successfully with instance id: mock-avs-default
INFO[0002] The installed node software has a plugin.
```

Notice the usage of:

* `--profile` to select the `option-returner` profile without prompt.
* `--no-prompt` to skip options prompts.

In this case, the `option-returner` profile uses all the default values. To set option values, use the `--option.<option-name>` dynamic flags. For instance:

```bash
$ eigenlayer install \
 --profile option-returner \
 --no-prompt \
 --option.main-port 8081 \
 https://github.com/NethermindEth/mock-avs
```

Output:

```bash
INFO[0000] Version not specified, using latest.
INFO[0000] Using version v3.1.0
INFO[0002] Installed successfully with instance id: mock-avs-default
INFO[0002] The installed node software has a plugin.
```

In this case, the `main-port` has a value of 8081 instead of the default value of 8080.

### From local directory

> THIS INSTALLATION METHOD IS INSECURE
> 

Installing from a local directory can be helpful for AVS developers who want to test Node Software packaging before releasing it to a public Git repository. To install an AVS Node Software from a local directory, use the `eigenlayer local-install` command. To illustrate local installation, let's clone the `mock-avs` to a local directory, and use it as a local package.

First, clone the `mock-avs` package:

> If you already have a local package, you can skip this step

```bash
git clone --branch v3.1.0 https://github.com/NethermindEth/mock-avs
```

Now we can install the package from the `mock-avs` directory with the following command:

```bash
eigenlayer local-install ./mock-avs --profile option-returner
```

Output:

```bash
WARN[0000] This command is insecure and should only be used for development purposes
INFO[0001] Installed successfully with instance id: mock-avs-default
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

```bash
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

```bash
INFO[0001] Running plugin with image eigen-plugin-mock-avs-default on network eigenlayer
INFO[0002]
AVS is up
```

This command will run the plugin container inside the AVS Docker network and remove it after execution.

The `--host` flag can be used to run the plugin container in the `host` network if the AVS Node Software is not running. To mount directories, files, and Docker volumes to the plugin container, use the `--volume` flag. All `plugin` command flags should be declared before the instance ID. To learn more about these options, run the `eigenlayer plugin --help` command.

### Plugin build args

If the plugin is built from a relative path inside the package or a remote context, the plugin image is built each time the plugin is executed. To pass build arguments to the plugin image build process, use the `--build-arg` flag, which is a map of key-value pairs. For example:

```bash
eigenlayer plugin \
 --build-arg arg1=value1 \
 --build-arg arg2=value2 \
 mock-avs-default \
 --port 8080
```

The `--build-arg` flag can be used multiple times to pass multiple build arguments. Should be declared before the instance ID to be recognized as a plugin build argument, and not as a plugin execution argument.

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