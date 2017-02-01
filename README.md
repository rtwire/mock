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
$ ./mock
```

This will run a local webserver that mimics RTWire's JSON endpoints at `http://localhost:8085/v1/mainnet/`. See our [documentation](https://rtwire.com/docs) for a description of available endpoints.

- The default port of  8085 can be changed by using the `-port` argument.
- The default authentication user name and password is `user` and `pass` respectively.
- Unlike the live API there is an extra endpoint at `http://localhost:[port]/v1/mainnet/addresses/[bitcoin address]`. This can be used to credit a public key hash address owned by the system.

## Example (Linux Based Systems)

The following example is taken from our API [walkthrough](https://rtwire.com/docs/walkthrough).

Start the mock service from the command line:
```bash
$ ./mock

2017/02/01 18:00:00 RTWire service running at http://localhost:8085/v1/mainnet/.
```

Either make the mock service run in the background (Ctrl+Z and then type `bg` on MacOS) or open a new command line window.


Create an account:
```bash
$ curl --user user:pass --header "Accept: application/json" --request POST --data "" http://localhost:8085/v1/mainnet/accounts/

{"type":"accounts","payload":[{"id":8674665223082153552,"balance":0}]}
```

Create an address assoicated with the account:
```bash
curl --user user:pass --header "Accept: application/json" --request POST --data "" http://localhost:8085/v1/mainnet/accounts/8674665223082153552/addresses/

{"type":"addresses","payload":[{"address":"1CXZFGn4jAdV7KTQvFL4FVSgSuCj2UECez"}]}
```

Send some funds to the address. (Note that this is a mock service specific endpoint. To fund a live account you would need to send a transaction to the address from a bitcoin wallet):
```bash
url --user user:pass --header "Content-Type: application/json" --header "Accept: application/json" --request POST --data '{"value": 2000}' http://localhost:8085/v1/mainnet/addresses/1CXZFGn4jAdV7KTQvFL4FVSgSuCj2UECez
```

Check that the funds have been sent to account `8674665223082153552`:
```bash
curl --user user:pass --header "Accept: application/json" --request GET --data '' http://localhost:8085/v1/mainnet/accounts/8674665223082153552
{"type":"accounts","payload":[{"id":8674665223082153552,"balance":2000}]}
```
