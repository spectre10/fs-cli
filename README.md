# fileshare-cli

fileshare-cli is CLI app written in Golang to transfer files via WebRTC protocol.

It is peer-to-peer (P2P), so there are no servers in middle. However, Google's STUN server is used to retrieve information about public address, the type of NAT clients are behind and the Internet side port associated by the NAT with a particular local port. (Transfer of files does not happen through Google servers.)

This information is used to setup Data Channel between clients.

https://github.com/spectre10/fileshare-cli/assets/72698233/db208f9d-7b94-4e58-9665-0a05e25e9b94


## Architecture

![WebRTC](https://github.com/spectre10/fileshare-cli/assets/72698233/5a13a571-51f6-400d-b534-492f9c38bc79)

# Installation

If you have Go installed, ([Install from here](https://go.dev/doc/install))

Add $GOPATH/bin to your $PATH. And then,

if you want the latest release version, then run this command,
```sh
go install github.com/spectre10/fileshare-cli@v0.1.2
```
or

if you want a specific release version, then run this command,
```sh
go install github.com/spectre10/fileshare-cli@vX.X.X
```
***

Alternatively, you can also download from GitHub Releases.

# Usage

To send a file,
```
fileshare-cli send --file <filepath>
```

To receive a file,
```
fileshare-cli receive
```

-----------------------------------
Currently only tested on Linux.
