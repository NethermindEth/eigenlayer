# EigenLayer CLI

EigenLayer CLI is used to manage core operator functionalities like local key management, operator registration and updates. 

- [EigenLayer CLI](#eigenlayer-cli)
  - [Dependencies](#dependencies)
  - [Install `eigenlayer` CLI using Go](#install-eigenlayer-cli-using-go)
  - [Install `eigenlayer` CLI from source](#install-eigenlayer-cli-from-source)
  - [Install `eigenlayer` CLI using a binary](#install-eigenlayer-cli-using-a-binary)
    - [Linux/amd64](#linuxamd64)
    - [Linux/arm64](#linuxarm64)
    - [Linux/amd64 (Ubuntu 20.04)](#linuxamd64-ubuntu-2004)
    - [Linux/arm64 (Ubuntu 20.04)](#linuxarm64-ubuntu-2004)
  - [Create and List Keys](#create-and-list-keys)
    - [Create keys](#create-keys)
    - [Import keys](#import-keys)
    - [List keys](#list-keys)
  - [Operator registration](#operator-registration)
    - [Sample config creation](#sample-config-creation)

## Supported Operating Systems
| Operating System | Architecture |
|------------------|--------------|
| Linux            | amd64        |
| Linux            | arm64        |


## Install `eigenlayer` CLI using a binary

The `eigenlayer` CLI tool versions are managed with GitHub releases. To install it, you can download the binary directly from the release assets manually, or by using the following command replacing the `<VERSION>` and `<ARCH>` with the proper values:

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/<VERSION>/eigenlayer-linux-<ARCH> --output eigenlayer
chmod +x eigenlayer
```

### Linux/amd64

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/v0.4.3/eigenlayer-linux-amd64 --output eigenlayer
chmod +x eigenlayer
```

### Linux/arm64

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/v0.4.3/eigenlayer-linux-arm64 --output eigenlayer
chmod +x eigenlayer
```

**_Note: If you are using `Ubuntu 20.04`, you can also use the below command to install the binary:_**
### Linux/amd64 (Ubuntu 20.04)

```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/v0.4.3/eigenlayer-linux-amd64-ubuntu-20-04 --output eigenlayer
chmod +x eigenlayer
```

### Linux/arm64 (Ubuntu 20.04)
```bash
curl -L https://github.com/NethermindEth/eigenlayer/releases/download/v0.4.3/eigenlayer-linux-arm64-ubuntu-20-04 --output eigenlayer
chmod +x eigenlayer
```

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

## Create or Import Keys
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
This outputs the public key and the ethereum address associated with the key. This will also be your operator address.
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

You can also list your created keys using

```bash
eigenlayer operator keys list
```

It will show all the keys created with this command with the public key

## Operator registration
>ECDSA and BLS keys are required for operator registration. 
You may choose to either [create](#create-keys) your own set of keys using the EigenLayer CLI (recommended for first time users) or [import](#import-keys) your existing keys (recommended for advanced users who already have keys created).

You can register your operator using the command below.

```bash
eigenlayer operator register operator-config.yaml
```

A sample yaml [operator-config.yaml](cli/operator/config/operator-config-example.yaml) is provided for reference. 
You can also create empty config files by using commands referred in [this section](#sample-config-creation). Fill in the required details to register the operator.

A public metadata url is required to register the operator. 
After creating and filling the [metadata](cli/operator/config/metadata-example.json) file, you can it to a publicly accessible location and give the url in the config file.
You are also required to upload the image of the operator to a publicly accessible location and give the url in the metadata file. We only support `.png` images for now. 

Make sure that if you use `local_keystore` as signer, you give the path to the keys created in above section.

After you complete the registration, you can check the registration status of your operator using

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