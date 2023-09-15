[![Go Reference](https://pkg.go.dev/badge/github.com/spectre10/fs-cli.svg)](https://pkg.go.dev/github.com/spectre10/fs-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/spectre10/fileshare-cli)](https://goreportcard.com/report/github.com/spectre10/fileshare-cli)


# fs-cli

fs-cli is multi-threaded CLI app written in Golang to transfer multiple files concurrently via WebRTC.

It is peer-to-peer (P2P), so there are no servers in middle. However, Google's STUN server is used to retrieve information about public address and the type of NAT clients are behind. (Transfer of files does not happen through Google servers.)

This information is used to setup a Peer Connection between clients. After connecting, each file is assigned a WebRTC Data-Channel for streaming. And the transfer happens concurrently over all Data-Channels. 

You can also find your public IP address via WebRTC. (See Usage Below)



https://github.com/spectre10/fs-cli/assets/72698233/fab8633b-af72-420c-9eff-3f91ada0eabc



Currently only tested on Linux.

## Architecture

![WebRTC](https://github.com/spectre10/fs-cli/assets/72698233/4c488af3-61e5-4e5f-9dc0-c5dfe528284e)


## Installation

If you have Go installed, ([Install from here](https://go.dev/doc/install))

Add $GOPATH/bin to your $PATH. And then,

if you want the latest release version, then run this command,
```sh
go install github.com/spectre10/fs-cli@latest
```
or

if you want a specific release version, then run this command,
```sh
go install github.com/spectre10/fs-cli@vX.X.X
```
***

Alternatively, you can also download from GitHub Releases.

## Usage

To send a file,
```
fs-cli send <filepath1> <filepath2> ... 
```

To receive a file,
```
fs-cli receive
```

To find your IP address,
```
fs-cli findip
```

## Future Steps

* (**On-going**) Add a web UI which can be hosted locally. (Single Binary.)

-----------------------------------
## References
* [pion/webrtc](https://github.com/pion/webrtc)
* [Antonito/gfile](https://github.com/Antonito/gfile)
