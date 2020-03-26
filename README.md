# Tenderly CLI

 [![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/Tenderly/tenderly-cli.svg?label=Latest%20Version)](https://github.com/Tenderly/tenderly-cli)

Tenderly CLI is a suite of development tools that allows you to debug, monitor and track the execution of your smart contracts.

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
    * [Login](#login)
    * [Init](#init)
    * [Push](#push)
    * [Export](#export)
    * [Local Proxy Debugging](#proxy-debugging)
    * [Check for updates](#check-for-updates)
    * [Version](#version)
    * [Who am I?](#who-am-i)
    * [Logout](#logout)

## Installation

### macOS

You can install the Tenderly CLI via the [Homebrew package manager](https://brew.sh/): 

```
brew tap tenderly/tenderly
brew install tenderly
```

Or if your prefer you can also install by using cURL and running our installation script:

```
curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-macos.sh | sh
```

### Linux

You can install the Tenderly CLI by using cURL and running our installation script:

```
curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-linux.sh | sh
```

### Windows

Go to the [release page](https://github.com/Tenderly/tenderly-cli/releases), download the latest version and put it somewhere in your `$PATH`.

### Updating

You can check the current version of the CLI by running:
```
tenderly version
```

To upgrade it via Homebrew:
```
brew upgrade tenderly
```

## Usage

### Login

The `login` command is used to authenticate the Tenderly CLI with your [Tenderly Dashboard](https://dashboard.tenderly.dev).

```
tenderly login
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --authentication-method | / | Pick the authentication method. Possible values are email or token |
| --email | / | The email used when authentication method is email |
| --password | / | The password used when authentication method is email |
| --token | / | The token used when authentication method is token |
| --force | false | Don't check if you are already logged in |
| --help | / | Help for login command |

### Init

The `init` command is used to connect your local project directory with a project in the [Tenderly Dashboard](https://dashboard.tenderly.dev).

```
tenderly init
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --project | / | The project name used for generating the configuration file |
| --create-project | false | Creates the project provided by the --project flag if it doesn't exist |
| --re-init | false | Force initializes the project if it was already initialized |
| --help | / | Help for init command |

### Push

The `push` command is used to add your contracts to the [Tenderly Dashboard](https://dashboard.tenderly.dev).

```
tenderly push
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --networks | / | A comma separated list of network ids to push |
| --tag | / | Optional tag used for filtering and referencing pushed contracts |
| --help | / | Help for push command |

#### Advanced usage

It is possible to push to multiple projects by editing the `tenderly.yaml` file and providing a map of projects and their networks. To do this remove the already provided `project_slug` property and replace it with the `projects` property like the example below;

```yaml
projects: # running tenderly push will push the smart contracts to all of the provided projects
  my-cool-project:
    networks:
    - "1" # mainnet
    - "42" # kovan
  my-other-project:
    # if the networks property is not provided or is empty the project will be pushed to all of the migrated networks
  company-account/my-other-project:
    # if you want to push to a shared project provide the full project identifier
    # the identifier can be found in you Tenderly dashboard under the projects name
```

### Export init

The `export init` subcommand helps define network in interactive mode.

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --project | / | The project name used for network configuration |
| --rpc-address | / | Rpc server address (example: 127.0.0.1:8545) |
| --forked-network | / | In case you forked a public network (example: mainnet) |
| --help | / | Help for export init command |

```
tenderly export init
```

### Export

The `export` command is used for local transaction debugging.

```
tenderly export {{transaction_hash}}
```

#### Command Arguments

| Name | Description |
| --- | --- |
| transaction hash | Hash of local transaction to debug |

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --export-network | / | Export network name located in tenderly yaml |
| --project | / | The project name |
| --rpc-address | / | Json rpc server address (example: 127.0.0.1:8545) |
| --forked-network | / | In case you forked a public network (example: mainnet) |
| --help | / | Help for export command |

#### Advanced usage

```yaml
exports: # running tenderly export will export local transaction to the provided project
  my-network:
    project_slug: my-cool-project
    rpc_address: 127.0.0.1:8545
    forked_network: mainnet
    chain_config:
      homestead_block: 0 # (default 0)
      eip150_block: 0 # (default 0)
      eip150_hash: 0x0 # (default 0x0)
      eip155_block: 0 # (default 0)
      eip158_block: 0 # (default 0)
      byzantium_block: 0 # (default 0)
      constantinople_block: 0 # (default 0)
      petersburg_block: 0 # (default 0)
      istanbul_block: 0 # (default 0)

  my-company-network:
    project_slug: company-account/my-other-project
    rpc_address: rpc.ethereum.company:8545
    # if you want to export to a shared project provide the full project identifier
    # the identifier can be found in you Tenderly dashboard under the projects name
```

### Verify

The `verify` command uploads your smart contracts and verifies them on [Tenderly](https://tenderly.dev).

```
tenderly verify
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --networks | / | A comma separated list of network ids to verify |
| --help | / | Help for verify command |

### Proxy Debugging

The proxy server is currently made to work with the [Truffle framework](https://truffleframework.com/) and requires the proxy to be run from the root of your Smart Contract project where the Truffle configuration is located.

```
// Example using Ganache defaults
tenderly proxy --target-port 7545
```

In your Truffle configuration, configure your local network config to point to the running proxy or create a new network for proxy debugging.

```
module.exports = {
    networks: {
        //...
        proxy: {
            host: "127.0.0.1",
            port: 9545,
            network_id: "*",
            gasPrice: 0
        },
        ganache: {
            host: "127.0.0.1",
            port: 7545,
            network_id: "*",
            gasPrice: 0
        },
        //...
    }
};
```

After setting up the network you can now call your Truffle commands just as before by changing the `--network` to the appropriate one.

```
$user > truffle exec ./scripts/test-scripts.js --network proxy
Using network 'proxy'.


Error: 0x0 Error: REVERT, execution stopped
	at require(square == original)
		in FailContract:5

...
```

Now when your transactions fail you can see the exact line of code in which the error occurred and the whole stacktrace by using our proxy command.

#### Note

You must run `truffle migrate --network proxy` first, so the contract information (address to source mapping) can be picked up by the `proxy` command from the `build` folder.

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --path | "./" | Path to the project build folder where your Truffle configuration is located |
| --proxy-host | "127.0.0.1" | Host on which the proxy will be listening |
| --proxy-port | "9545" | Port on which the proxy will be listening |
| --target-host | "127.0.0.1" | Target host of your Blockchain RPC |
| --target-port | "8545" | Target port of your Blockchain RPC |
| --target-schema | "http" | Blockchain RPC protocol |
| --write-config | / | Write proxy settings to the project configuration file |
| --help | / | Help for proxy command |

### Check for updates

The `update-check` command checks if there is a new version of the Tenderly CLI and gives update instructions and changelog information.

### Version

The `version` command prints out the current version of the Tenderly CLI.

```
tenderly version
```

### Who am I?

The `whomai` command prints out basic information about the currently logged in account

```
tenderly whoami
```

### Logout

The `logout` command disconnects your local Tenderly CLI from your [Tenderly Dashboard](https://dashboard.tenderly.dev)

```
tenderly logout
```

### Global Flags

In addition to command specific flags, the following flags can be passed to any command

| Flag | Default | Description |
| --- | --- | --- |
| --debug | false | Turn on debug level logging |
| --global-config | config | Global configuration file name (without the extension) |
| --project-config | tenderly | Project configuration file name (without the extension) |
| --project-dir | "./" | The directory in which your Truffle project resides |

## Report Bugs / Feedback

We look forward to any feedback you want to share with us or if you're stuck with a problem you can contact us at [support@tenderly.app](mailto:support@tenderly.app).

You can also join our [Discord server](https://discord.gg/fBvDJYR) or create an Issue in the Github repository.

-----

Made with ♥ by [Tenderly](https://tenderly.dev)
