# Tenderly CLI

 [![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/Tenderly/tenderly-cli.svg?label=Latest%20Version)](https://github.com/Tenderly/tenderly-cli)

Tenderly CLI is a suite of development tools that allows you to debug, monitor and track the execution of your smart contracts.

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
    * [Login](#login)
    * [Init](#init)
    * [Push](#push)
    * [Export setup](#export-init)
    * [Export local transactions to Tenderly](#export)
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

The `login` command is used to authenticate the Tenderly CLI with your [Tenderly Dashboard](https://dashboard.tenderly.co).

```
tenderly login
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --authentication-method | / | Pick the authentication method. Possible values are email or access-key |
| --email | / | The email used when authentication method is email |
| --password | / | The password used when authentication method is email |
| --token | / | The token used when authentication method is token |
| --force | false | Don't check if you are already logged in |
| --help | / | Help for login command |

### Init

The `init` command is used to connect your local project directory with a project in the [Tenderly Dashboard](https://dashboard.tenderly.co).

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

The `push` command is used to add your contracts to the [Tenderly Dashboard](https://dashboard.tenderly.co).

Note that the `push` command is used **only** for adding contracts that are deploy to a public network. For local networks see
the [export command](#export).

```
tenderly push
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --networks | / | A comma separated list of network ids to push |
| --tag | / | Optional tag used for filtering and referencing pushed contracts |
| --project-slug | / | Optional project slug used to pick only one project to push (see advanced usage) |
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

In order to use the [tenderly export](#export) command you need to define a configuration file
(which is described in more detail in the [export command](#export) advanced usage section).

```
tenderly export init
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --project | / | The project name used for network configuration |
| --rpc | / | Rpc server address (example: 127.0.0.1:8545) |
| --forked-network | / | In case you forked a public network (example: mainnet) |
| --help | / | Help for export init command |

### Export

The `export` command can be used to access transaction debugging tooling available at https://dashboard.tenderly.co/ but for local transactions.


Use the
[Transaction Overview](https://dashboard.tenderly.co/tx/main/0x70f28ce44bd58034ac18bec9eb1603350d50e020e4c2cf0b071837699ea1cdb1),
[Human-Readable Stack-Traces](https://dashboard.tenderly.co/tx/main/0x30bc65375b2e2b56f97706bccba9b21bc8763cc81a0262351b3373ce49f60ea7),
[Debugger](https://dashboard.tenderly.co/tx/main/0x70f28ce44bd58034ac18bec9eb1603350d50e020e4c2cf0b071837699ea1cdb1/debugger),
[Gas Profiler](https://dashboard.tenderly.co/tx/main/0x70f28ce44bd58034ac18bec9eb1603350d50e020e4c2cf0b071837699ea1cdb1/gas-usage),
[Decoded Events](https://dashboard.tenderly.co/tx/main/0x70f28ce44bd58034ac18bec9eb1603350d50e020e4c2cf0b071837699ea1cdb1/logs) and [State](https://dashboard.tenderly.co/tx/main/0x70f28ce44bd58034ac18bec9eb1603350d50e020e4c2cf0b071837699ea1cdb1/state-diff)
to boost your local development productivity.
```
tenderly export {{transaction_hash}}
```

#### Command Arguments

| Name | Description |
| --- | --- |
| transaction hash | Hash of the local transaction to debug |

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --export-network | / | The name of the exported network in the configuration file |
| --project | / | The project in which the exported transactions will be stored |
| --rpc | 127.0.0.1:8545 | The address and port of the local rpc node |
| --forked-network | / | Optional name of the network which you are forking locally. Can be one of Mainnet, Goerli, Kovan, Ropsten, Rinkeby, xDai |
| --protocol | / | Specify the protocol used for the rpc node. By default `wss`, `https`, `ws`, `http` are tried in that order |
| --help | / | Help for export command |
| --force| false | Export the transaction regardless of gas mismatch|

#### Advanced usage

If your local node has different blocks defined for hardforks or you want to generate the configuration file yourself,
you can find the example bellow:

```yaml
exports: # running tenderly export will export local transaction to the provided project
  my-network:
    project_slug: my-cool-project
    rpc_address: 127.0.0.1:8545
    protocol: http
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
      berlin_block: 0 # (default 0)
      london_block: 0 # (default 0)

  my-company-network:
    project_slug: company-account/my-other-project
    rpc_address: rpc.ethereum.company:8545
    # if you want to export to a shared project provide the full project identifier
    # the identifier can be found in you Tenderly dashboard under the projects name
```

### Verify

The `verify` command uploads your smart contracts and verifies them on [Tenderly](https://tenderly.co).

```
tenderly verify
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --networks | / | A comma separated list of network ids to verify |
| --help | / | Help for verify command |

### Check for updates

The `update-check` command checks if there is a new version of the Tenderly CLI and gives update instructions and changelog information.

### Version

The `version` command prints out the current version of the Tenderly CLI.

```
tenderly version
```

### Who am I?

The `whoami` command prints out basic information about the currently logged in account

```
tenderly whoami
```

### Logout

The `logout` command disconnects your local Tenderly CLI from your [Tenderly Dashboard](https://dashboard.tenderly.co)

```
tenderly logout
```

### Proxy Debugging

The proxy command is deprecated in favor of the [export](#export) command.

### Global Flags

In addition to command specific flags, the following flags can be passed to any command

| Flag | Default | Description |
| --- | --- | --- |
| --debug | false | Turn on debug level logging |
| --output | text | Which output mode to use: text or json. If not provided. text output will be used. |
| --global-config | config | Global configuration file name (without the extension) |
| --project-config | tenderly | Project configuration file name (without the extension) |
| --project-dir | "./" | The directory in which your Truffle project resides |

## Report Bugs / Feedback

We look forward to any feedback you want to share with us or if you're stuck with a problem you can contact us at [support@tenderly.co](mailto:support@tenderly.co).

You can also join our [Discord server](https://discord.gg/fBvDJYR) or create an Issue in the Github repository.

-----

Made with ♥ by [Tenderly](https://tenderly.co)
