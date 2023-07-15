# fileshare-cli

fileshare-cli is multi-threaded CLI app written in Golang to transfer multiple files concurrently via WebRTC protocol.

It is peer-to-peer (P2P), so there are no servers in middle. However, Google's STUN server is used to retrieve information about public address, the type of NAT clients are behind and the Internet side port associated by the NAT with a particular local port. (Transfer of files does not happen through Google servers.)

This information is used to setup Data Channel between clients.

You can also find your public IP address via WebRTC. (See Usage)

https://github.com/spectre10/fileshare-cli/assets/72698233/bc1e2863-1b17-4ccd-b7e6-aae743844676

## Architecture

![webrtc](https://github.com/spectre10/fileshare-cli/assets/72698233/d6e92b4e-ceea-46f7-83d1-cb994a75774f)



# Installation

If you have Go installed, ([Install from here](https://go.dev/doc/install))

Add $GOPATH/bin to your $PATH. And then,

if you want the latest release version, then run this command,
```sh
go install github.com/spectre10/fileshare-cli@latest
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
fileshare-cli send <filepath1> <filepath2> ... 
```

To receive a file,
```
fileshare-cli receive
```

To find your IP address,
```
fileshare-cli findip
```

-----------------------------------
Currently only tested on Linux.
