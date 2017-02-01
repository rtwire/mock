# mock

mock is a mock implementation of the RTWire HTTP endpoints that allows local unit and integration testing. Documentation of the endpoints can be found at [https://rtwire.com/docs](https://rtwire.com/docs).

## Requirements
[Go](http://golang.org) 1.7 or newer.


## Installation

### Build From Source – (Linux Based Systems)

- Install Go version 1.7 or newer according to the installation instructions here: [https://golang.org/doc/install](https://golang.org/doc/install).

- Ensure GOPATH is set correctly. It's usually set to a location such as your home directory:

```bash
$ export GOPATH=[directory you choose]
```

- Run the following commands to download mock, all its dependencies and install:

```bash
$ export PATH=$GOPATH/bin:$PATH
$ go get -u github.com/Masterminds/glide
$ git clone https://github.com/rtwire/mock $GOPATH/src/github.com/rtwire/mock
$ cd $GOPATH/src/github.com/rtwire/mock
$ glide install
$ go install
```
- mock will now be installed in ```$GOPATH/bin```.

### Download Binaries

- Instead of building from source you can also download a precompiled binary for the latest release found at [https://github.com/rtwire/mock/releases](https://github.com/rtwire/mock/releases).

## Running

You can run mock as follows:

```bash
$ ./mock -port [port]
```

This will run a local webserver at the port you choose. The default port is 8085 if none is choosen. RTWire API endpoints accessable at http://localhost:[port]/v1/mainnet/. See our [documentation](https://rtwire.com/docs) for a description of available endpoints.

Note that for unit/integration testing purposes there is an extra HTTP POST endpoint at http://localhost:[port]/v1/mainnet/addresses/[bitcoin address]. This can be used to simulate crediting an address and its respective account without having to send a transaction on the actual bitcoin network.
