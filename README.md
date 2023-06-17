# fileshare-cli

fileshare-cli is CLI app written in Golang to transfer files via WebRTC protocol.

It is peer-to-peer (P2P), so there are no servers in middle. However, Google's STUN server is used to retrieve information about public address, the type of NAT clients are behind and the Internet side port associated by the NAT with a particular local port. (Transfer of files does not happen through Google servers.)

This information is used to setup Data Channel between clients.

https://github.com/spectre10/fileshare-cli/assets/72698233/86917b69-1137-4496-9f4c-3dacdccd31ae

## Architecture

![WebRTC](https://github.com/spectre10/fileshare-cli/assets/72698233/5a13a571-51f6-400d-b534-492f9c38bc79)

# Installation

If you have Go installed, ([Install from here](https://go.dev/doc/install))

you want a stable release version, then run this command,

`$ go install github.com/spectre10/fileshare-cli@v0.1.0`

you want the latest git version, then run this command,

`$ go install github.com/spectre10/fileshare-cli@latest`


Alternatively, you can download from GitHub Releases.

# Usage

To send a file,
`$ fileshare-cli send --file <filepath>`

To receive a file,
`$ fileshare-cli receive --file <filepath>`
