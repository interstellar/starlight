![Header image](doc/images/github-readme-header@2x.png)

# Starlight

Starlight is a demo implementation of payment channels on Stellar.

Payment channels allow parties to transact privately, instantly, and securely, while paying zero fees. They are also a first step toward constructing and connecting to payment channel networks like [Lightning](https://lightning.network/) and [Interledger](https://interledger.org/).

This release includes the Starlight wallet, a user interface that lets you create bilateral payment channels and use them to transact in lumens (the native asset of the Stellar protocol) on the Stellar testnet.

You can learn more about the Starlight protocol by reading [the specification](doc/Protocol.md), or you can try out the wallet by following [the installation instructions](#installation) below.

If you encounter any bugs, want to request a feature, or have any questions about the software, you can open an [issue](https://github.com/interstellar/starlight/issues) or [pull request](https://github.com/interstellar/starlight/pulls) on GitHub, or talk to us in the #starlight channel on the [Interstellar Slack](https://slack.interstellar.com/).

**Please note that this software is still in development and will likely contain multiple security and functional bugs and issues. This code is for testing only on the Stellar testnet. It is not compatible with the Stellar mainnet. Attempting to modify it for use on the Stellar mainnet may result in permanent XLM loss. Any use of this software is strictly at your own risk, and we will not be liable for any losses or damages.**

- [Starlight](#starlight)
    - [Demo](#demo)
    - [Setup](#setup)
    - [Tutorial](#tutorial)
    - [Development](#development)
    - [Roadmap](#roadmap)

## Demo

![Demo](doc/images/demo.gif)

## Setup

### Installation

You can download the appropriate binary for your platform from the [releases](https://github.com/interstellar/starlight/releases) page. Move the binary to somewhere on your executable path (for example, with `sudo mv starlightd /usr/local/bin`).

Alternatively, instructions for installing from source are [below](#development).

You can run a Starlight instance using the `starlightd` command.
This will set up a data directory, called `starlight-data`, in your current directory.
You can then open the wallet by going to [http://localhost:7000](http://localhost:7000) in your browser.

If, at any time, you want to reset your agent completely, you can clear your data directory:

```sh
$ rm -rf starlight-data
```

### Running a second instance on the same computer

To try out `starlightd` for the first time,
you'll want to run two instances on the same computer and create a payment channel between them,
as described in [the tutorial](#tutorial).
You can run a second instance on your computer by listening on a different port,
and using a separate data directory.

```sh
$ starlightd --listen=localhost:7001 --data=starlight-data-2
```

To log in to this wallet, go to [http://localhost:7001](http://localhost:7001), using an incognito or private window (to avoid logging out your other session).

### Connecting to instances on other computers

If you are running your Starlight instance on your personal computer and want to connect it with instances on other computers or servers, you'll need to give your instance a publicly-accessible URL. One way to do so is by using a service like [Serveo](https://serveo.net) or [ngrok](https://ngrok.com). For example, you can run the following command:

```sh
$ ssh -R 80:localhost:7000 serveo.net
```

Your Starlight address will then be served on a subdomain of serveo.net (so your Stellar address will be something like alice\*something.serveo.net).

### Running an instance on AWS

Alternatively, you can run your Starlight instance on a cloud computing platform like Amazon Web Services or DigitalOcean. This more closely resembles how future production versions of Starlight would likely be hosted.

You can find instructions for setting up a Starlight instance on AWS [here](infra/services/starlight/README).

## Tutorial

Start by [installing](#installation) `starlightd`, setting up [two instances](#running-a-second-instance-on-the-same-computer) locally, and opening two browser windows to [http://localhost:7000](http://localhost:7000) and [http://localhost:7001](http://localhost:7001) (one of which should be in a private or incognito window, to prevent the sessions from interfering with each other).

Configure each wallet, picking "alice" and "bob" as the respective usernames.

### Wallet

Starlight provides a simple lumen wallet, which manages an account that is funded with 10,000 testnet lumens upon setup.

The wallet gives you a Stellar address, e.g., "alice\*localhost:7000".

You can use this wallet to make on-network payments to users' Stellar addresses (i.e., alice\*stellar.org) or their Stellar account IDs (e.g., GAIH3ULLFQ4DGSECF2AR555KZ4KNDGEKN4AFI4SU2M7B43MGK3QJZNSR).

Try having Alice send a 100 XLM payment to Bob.

### Channels

The core feature of Starlight is payment channels. Payment channels allow two parties to make payments to each other using free, private, and secure off-network transactions.

Try having Alice create a 500 XLM payment channel with Bob (bob\*localhost:7001).

The channel will take a few seconds to open. Once the channel is open, try having Alice send 100 XLM to Bob, then try having Bob send 50 XLM back to Alice. Note that since these payments happened in the payment channel rather than on the public network, they were almost instant, and the parties paid no fee.

### Channel capacity

Examine the Capacity graph, which shows how much each party can currently send and receive in the channel. The total capacity of the channel is limited to the amount that has been deposited in it. While the parties can make payments back and forth as long as they like, no party can make a payment that would exceed the channel's capacity.

The party that created the channel can deposit additional funds into the channel. This moves funds from their wallet account to their balance in the channel, and increases the total capacity of the channel. Try having Alice deposit 500 XLM into the channel, by clicking Deposit on the channel page.

### Closing a channel

The funds that are locked in this channel can be paid back and forth between Alice and Bob instantly. However, if Alice or Bob want to make payments to anyone else, or use those lumens in any other channel, they will need to withdraw the funds by closing the channel.

Try having Bob close the channel by clicking Close on the channel page. After a few seconds, the channel should close, and the parties' funds should be withdrawn to their wallet accounts.

This worked because Bob's instance was online, and it automatically cooperated with the channel close request. If Bob's instance was offline or did not cooperate, Alice would need to "force close" the channel, which would mean that there would be some delay before she would receive her funds.

## Development

### Build starlightd from source

To build the `starlightd` agent from source,
you'll need to install [Go](https://golang.org/doc/install) version 1.11,
and set up a properly configured [$GOPATH](https://github.com/golang/go/wiki/GOPATH) directory,
with `$GOPATH/bin` added to your PATH:

```
export PATH=$PATH:$GOPATH/bin
```

Install `starlightd` from its GitHub repository.

```sh
$ go get github.com/interstellar/starlight/...
```

You can now run the command `starlightd` from anywhere. This will create a data directory called `starlight-data` in your current directory, and will run a wallet, which you can access at http://localhost:7000.

### Build wallet from source

When running the `starlightd` binary, it automatically downloads and serves the latest version of the front-end wallet.

If you want to make changes to the wallet source code, you can rebuild and run it independently:

```sh
$ cd $GOPATH/github.com/interstellar/starlight/wallet
$ npm install
$ npm start
```

The wallet should now be running at port 5000:

```sh
$ open http://localhost:5000
```

To run a second development wallet on port 5001, connecting to a `starlightd` running on port 7001:

```sh
$ PORT=5001 STARLIGHTD_URL=http://localhost:7001 npm start
```

### Running tests

The Starlight project has unit tests and integration tests for the
starlight server and the wallet frontend.

To run the Starlight server unit tests:

```sh
$ cd $I10R/starlight
$ go test -short ./...
```

To run the Starlight wallet unit tests:

```sh
$ cd $I10R/starlight/wallet
$ ./bin/tests
```

To run the Starlight integration tests:

```sh
$ cd $I10R/starlight/starlighttest
$ go test
```

## Roadmap

Starlight is under active development at Interstellar. Our top priorities for the coming year include:

- Stabilization and final specification of protocol and API
- Channels for non-native assets
- Cross-channel atomic payments
- Cross-currency atomic payments
- Compatibility with Interledger and Lightning
- Peer-to-peer connectivity
- Stellar mainnet launch

To learn more about these projects, or to let us know what features you would most like to see, you can join the discussion in the #starlight channel on [the Interstellar Slack](https://slack.interstellar.com).
