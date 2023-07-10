package lib

import (
	"os"

	"github.com/pion/webrtc/v3"
)

type Metadata struct {
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

type Document struct {
	*Metadata
	MetadataDone bool
	File         *os.File
	Packet       []byte
	DC           *webrtc.DataChannel
	DCdone       chan struct{}
}
