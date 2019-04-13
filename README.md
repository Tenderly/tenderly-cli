# Tenderly CLI

![CircleCI token](https://img.shields.io/circleci/token/d03a8a252d1d376e478938b24522714ca678cfcc/project/github/Tenderly/tenderly-cli.svg?label=Build&logo=circleci) [![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/Tenderly/tenderly-cli.svg?label=Latest%20Version)](https://github.com/Tenderly/tenderly-cli)

Tenderly CLI is a suite of development tools that allows you to debug, monitor and track the execution of your smart contracts.

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

#### Command Flags

| Flag | Default | Description |
| --- | --- | --- |
| --path | "./" | Path to the project build folder where your Truffle configuration is located. |
| --proxy-host | "127.0.0.1" | Host on which the proxy will be listening |
| --proxy-port | "9545" | Port on which the proxy will be listening |
| --target-host | "127.0.0.1" | Target host of your Blockchain RPC |
| --target-port | "8545" | Target port of your Blockchain RPC |
| --target-schema | "http" | Blockchain RPC protocol |
| --help | / | Help for proxy command |

## Report Bugs / Feedback

We look forward to any feedback you want to share with us or if you're stuck with a problem you can contact us at [support@tenderly.app](mailto:support@tenderly.app).

You can also join our [Discord server](https://discord.gg/fBvDJYR) or create an Issue in the Github repository.

-----

Made with ♥ by [Tenderly](https://tenderly.dev)
