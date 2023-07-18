package lib

import (
	"os"

	"github.com/pion/webrtc/v3"
)

// This struct is sent first to receiver as JSON to ask for consent and display the stats.
type Metadata struct {
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

// This struct is essentially a file and its associated information (including the WebRTC datachannel).
type Document struct {
	*Metadata
	MetadataDone bool
	File         *os.File
	Packet       []byte
	DC           *webrtc.DataChannel
	DCdone       chan struct{}
	DCclose      chan struct{}
	StartTime    int64
}
