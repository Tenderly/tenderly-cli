# Tenderly CLI

[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/Tenderly/tenderly-cli.svg?label=Latest%20Version)](https://github.com/Tenderly/tenderly-cli)

Tenderly CLI is a suite of development tools that allows you to debug, monitor and track the execution of your smart
contracts. 

For smart contract verification in Tenderly, follow the [Foundry verification guide](https://docs.tenderly.co/contract-verification/foundry) and [Hardhat verification guide](https://docs.tenderly.co/contract-verification/hardhat).

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
    * [Login](#login)
    * [Init](#init)
    * [Push](#push)
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

Or if you prefer you can also install by using cURL and running our installation script:

```
curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-macos.sh | sh
```

### Linux

You can install the Tenderly CLI by using cURL and running our installation script.

With `root` privileges user:
```
curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-linux.sh | sh
```

Or with `sudo` user:
```
curl https://raw.githubusercontent.com/Tenderly/tenderly-cli/master/scripts/install-linux.sh | sudo sh
```


### Windows

Go to the [release page](https://github.com/Tenderly/tenderly-cli/releases), download the latest version and put it
somewhere in your `$PATH`.

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

The `login` command is used to authenticate the Tenderly CLI with
your [Tenderly Dashboard](https://dashboard.tenderly.co).

```
tenderly login
```

#### Command Flags

| Flag | Default | Description                                                            |
| --- | --- |------------------------------------------------------------------------|
| --authentication-method | / | Pick the authentication method. Possible values are email or access-key |
| --email | / | The email used when authentication method is email                     |
| --password | / | The password used when authentication method is email                  |
| --access-key | / | The token used when authentication method is access-key         |
| --force | false | Don't check if you are already logged in                               |
| --help | / | Help for login command                                                 |

### Init

The `init` command is used to connect your local project directory with a project in
the [Tenderly Dashboard](https://dashboard.tenderly.co).

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

### Example how to initialize tenderly project

For Tenderly CLI to work you need to have a `deployments` directory inside your project. You can generate that one
using [hardhat-tenderly](https://github.com/Tenderly/hardhat-tenderly#readme.)

1. To install hardhat-tenderly run.

```bash
npm install --save-dev @tenderly/hardhat-tenderly
```

2. Add the following statement to your `hardhat.config.js`:

```js
require("@tenderly/hardhat-tenderly");
```

Or, if you are using typescript:

```js
import "@tenderly/hardhat-tenderly"
```

3. Then you need to call it from your scripts (using ethers to deploy a contract):

```js
const Greeter = await ethers.getContractFactory("Greeter");
const greeter = await Greeter.deploy("Hello, Hardhat!");

await greeter.deployed()

await hre.tenderly.persistArtifacts({
    name: "Greeter",
    address: greeter.address,
})
```

`persistArtifacts` accept variadic parameters:

```js
const contracts = [
    {
        name: "Greeter",
        address: "123"
    },
    {
        name: "Greeter2",
        address: "456"
    }
]

await hre.tenderly.persistArtifacts(...contracts)
```

4. Run: `npx hardhat compile` to compile contracts
5. Run: `npx hardhat node --network hardhat` to start a local node
6. Run: `npx hardhat run scripts/sample-script.js --network localhost` to run a script
7. And at the end now when `deployments` directory was built you can run `tenderly init`

### Push

If you are using Hardhat, take a look at [docs](https://docs.tenderly.co/monitoring/smart-contract-verification/verifying-contracts-using-the-tenderly-hardhat-plugin) instead of using this command.

The `push` command is used to add your contracts to the [Tenderly Dashboard](https://dashboard.tenderly.co).

Note that the `push` command is used **only** for adding contracts that are deployed to a public network.

```
tenderly contracts push
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --networks | / | A comma separated list of network ids to push |
| --tag | / | Optional tag used for filtering and referencing pushed contracts |
| --project-slug | / | Optional project slug used to pick only one project to push (see advanced usage) |
| --help | / | Help for push command |

#### Advanced usage

It is possible to push to multiple projects by editing the `tenderly.yaml` file and providing a map of projects and
their networks. To do this remove the already provided `project_slug` property and replace it with the `projects`
property like the example below;

```yaml
projects: # running tenderly push will push the smart contracts to all of the provided projects
  my-cool-project:
    networks:
      - "1" # mainnet
      - "5" # goerli
  my-other-project:
  # if the networks property is not provided or is empty the project will be pushed to all of the migrated networks
  company-account/my-other-project:
  # if you want to push to a shared project provide the full project identifier
  # the identifier can be found in your Tenderly dashboard under the projects name
```

### Verify

The `verify` command uploads your smart contracts and verifies them on [Tenderly](https://tenderly.co).

```
tenderly contracts verify
```

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --networks | / | A comma separated list of network ids to verify |
| --help | / | Help for verify command |

### Check for updates

The `update-check` command checks if there is a new version of the Tenderly CLI and gives update instructions and
changelog information.

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

We look forward to any feedback you want to share with us or if you're stuck with a problem you can contact us
at [support@tenderly.co](mailto:support@tenderly.co).

-----

Made with ♥ by [Tenderly](https://tenderly.co)
